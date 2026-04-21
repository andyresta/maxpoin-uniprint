// Package printer menangani deteksi printer fisik (spooler OS dan
// Bluetooth) serta pengiriman job pencetakan ke printer tujuan.
// Implementasi cross-platform untuk Windows / Linux / macOS.
package printer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"printbridge/internal/config"
	"printbridge/internal/escpos"
	"printbridge/internal/imageproc"
)

// Konstanta tipe printer untuk hasil deteksi.
const (
	TypeSpooler   = "spooler"
	TypeBluetooth = "bluetooth"
)

// Konstanta status printer.
const (
	StatusReady   = "ready"
	StatusOffline = "offline"
)

// Info mendeskripsikan satu printer yang terdeteksi pada sistem.
type Info struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	// Port adalah identifier transport untuk printer Bluetooth
	// (contoh "COM5" di Windows atau "/dev/rfcomm0" di Linux).
	// Untuk printer spooler, field ini bisa kosong.
	Port string `json:"port,omitempty"`
}

// minRefreshInterval adalah jeda minimum antar dua refresh penuh.
// Permintaan refresh yang datang lebih cepat dari interval ini langsung
// mengembalikan data cache tanpa menjalankan ulang perintah OS.
const minRefreshInterval = 3 * time.Second

// Manager menyimpan hasil deteksi printer terbaru dan menyediakan
// method dispatch (PrintText/PrintImage) yang aman dipakai concurrent.
type Manager struct {
	mu            sync.RWMutex
	printers      []Info
	lastRefreshed time.Time // kapan Refresh() terakhir dijalankan
}

// NewManager membuat Manager kosong; daftar printer harus diisi
// melalui Refresh() (biasanya dijalankan di goroutine background saat startup).
func NewManager() *Manager {
	return &Manager{printers: []Info{}}
}

// detectionResult adalah container hasil satu goroutine deteksi.
type detectionResult struct {
	printers []Info
	err      error
}

// Refresh melakukan deteksi ulang seluruh printer (spooler + Bluetooth)
// secara PARALEL menggunakan dua goroutine, sehingga total waktu tunggu
// adalah max(waktu_spooler, waktu_bluetooth) bukan penjumlahannya.
//
// Jika Refresh() sudah dipanggil kurang dari minRefreshInterval yang lalu,
// fungsi ini langsung kembali dengan data cache tanpa menjalankan ulang
// perintah OS, mencegah loading lambat akibat klik berulang.
//
// Context dipakai untuk timeout/cancel keseluruhan proses deteksi.
func (m *Manager) Refresh(ctx context.Context) error {
	// Cek apakah data cache masih cukup segar (debounce).
	m.mu.RLock()
	tooSoon := !m.lastRefreshed.IsZero() && time.Since(m.lastRefreshed) < minRefreshInterval
	m.mu.RUnlock()
	if tooSoon {
		return nil
	}

	spoolCh := make(chan detectionResult, 1)
	btCh := make(chan detectionResult, 1)

	// Jalankan deteksi spooler dan Bluetooth secara paralel.
	go func() {
		p, err := detectSpoolerPrinters(ctx)
		spoolCh <- detectionResult{p, err}
	}()
	go func() {
		p, err := detectBluetoothPrinters(ctx)
		btCh <- detectionResult{p, err}
	}()

	// Tunggu kedua goroutine selesai (atau context timeout).
	var sr, br detectionResult
	select {
	case sr = <-spoolCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case br = <-btCh:
	case <-ctx.Done():
		return ctx.Err()
	}

	combined := make([]Info, 0, len(sr.printers)+len(br.printers))
	combined = append(combined, sr.printers...)
	combined = append(combined, br.printers...)

	m.mu.Lock()
	m.printers = combined
	m.lastRefreshed = time.Now()
	m.mu.Unlock()

	// Gabungkan error tanpa membatalkan jika salah satu sumber gagal.
	switch {
	case sr.err != nil && br.err != nil:
		return fmt.Errorf("deteksi spooler & bluetooth gagal: %v ; %v", sr.err, br.err)
	case sr.err != nil:
		return fmt.Errorf("deteksi spooler gagal: %w", sr.err)
	case br.err != nil:
		return fmt.Errorf("deteksi bluetooth gagal: %w", br.err)
	}
	return nil
}

// List mengembalikan salinan daftar printer yang sudah terdeteksi.
func (m *Manager) List() []Info {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Info, len(m.printers))
	copy(out, m.printers)
	return out
}

// Find mencari printer berdasarkan nama (case-insensitive).
// Mengembalikan info dan true bila ditemukan.
func (m *Manager) Find(name string) (Info, bool) {
	target := strings.ToLower(strings.TrimSpace(name))
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.printers {
		if strings.ToLower(p.Name) == target {
			return p, true
		}
	}
	return Info{}, false
}

// TextJob merepresentasikan permintaan cetak teks dari API.
type TextJob struct {
	Text         string
	LogoBitmap   *RasterImage // opsional; sudah dalam bentuk raster monochrome
	LogoPosition string       // "left" | "center" | "right"
	Copies       int
	CutPaper     bool
	Encoding     string
	// Mode menentukan cara pembentukan payload:
	//   "thermal"   → ESC/POS (default bila kosong)
	//   "dotmatrix" → plain text tanpa ESC/POS
	// Bila kosong, diambil dari config.Config.PrinterMode.
	Mode string
}

// ImageJob merepresentasikan permintaan cetak gambar dari API.
type ImageJob struct {
	Bitmap   *RasterImage
	Copies   int
	CutPaper bool
	// Mode: "thermal" didukung; "dotmatrix" akan menghasilkan error
	// karena printer dot matrix tidak mendukung bitmap raster ESC/POS.
	Mode string
}

// RasterImage menampung data raster monochrome yang sudah siap dicetak
// melalui perintah ESC/POS GS v 0.
type RasterImage struct {
	Data       []byte
	WidthBytes int
	HeightPx   int
}

// resolveMode menentukan mode aktif untuk sebuah job. Jika job.Mode
// sudah diisi secara eksplisit, nilai itu yang dipakai. Jika kosong,
// fallback ke cfg.PrinterMode. Ultimate default adalah "thermal".
func resolveMode(jobMode, cfgMode string) string {
	m := strings.ToLower(strings.TrimSpace(jobMode))
	if m == config.PrinterModeThermal || m == config.PrinterModeDotMatrix {
		return m
	}
	m = strings.ToLower(strings.TrimSpace(cfgMode))
	if m == config.PrinterModeThermal || m == config.PrinterModeDotMatrix {
		return m
	}
	return config.PrinterModeThermal
}

// PrintText membentuk payload cetak teks lalu mengirimnya ke printer.
//
// Mode "thermal": payload ESC/POS dengan dukungan logo raster, alignment,
// dan perintah cut paper.
//
// Mode "dotmatrix": payload teks polos (tanpa byte ESC/POS) yang sesuai
// untuk printer dot matrix maupun spooler plain-text. Logo dan cut paper
// diabaikan.
func (m *Manager) PrintText(ctx context.Context, cfg config.Config, job TextJob) error {
	if strings.TrimSpace(cfg.DefaultPrinter) == "" {
		return errors.New("default printer belum dikonfigurasi")
	}
	info, ok := m.Find(cfg.DefaultPrinter)
	if !ok {
		// Tetap coba kirim ke spooler sesuai nama; deteksi mungkin tertinggal.
		info = Info{Name: cfg.DefaultPrinter, Type: cfg.PrinterType, Status: StatusReady}
	}

	if job.Copies < 1 {
		job.Copies = 1
	}
	if job.Copies > 10 {
		job.Copies = 10
	}

	mode := resolveMode(job.Mode, cfg.PrinterMode)

	var payload []byte
	if mode == config.PrinterModeDotMatrix {
		// Plain text: kirim byte mentah tanpa ESC/POS apa pun.
		// Setiap copy dipisahkan form-feed (0x0C) agar printer
		// memajukan ke baris baru yang bersih.
		text := job.Text
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		payload = []byte(text)
	} else {
		// Thermal (ESC/POS)
		b := escpos.New().Init()
		if job.LogoBitmap != nil {
			alignByte := alignmentByteFromString(job.LogoPosition)
			b.Align(alignByte)
			if err := b.PrintRasterBitmap(job.LogoBitmap.Data, job.LogoBitmap.WidthBytes, job.LogoBitmap.HeightPx); err != nil {
				return fmt.Errorf("logo raster invalid: %w", err)
			}
			b.LineFeed()
		}
		b.Align(escpos.AlignLeft)
		text := job.Text
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		b.Text(text)
		b.FeedLines(2)
		if job.CutPaper {
			b.CutPaper(true)
		}
		payload = b.Bytes()
	}

	for i := 0; i < job.Copies; i++ {
		if err := sendToPrinter(ctx, info, payload); err != nil {
			return fmt.Errorf("gagal kirim ke printer (copy %d): %w", i+1, err)
		}
	}
	return nil
}

// PrintImage mengirim gambar raster ESC/POS ke printer thermal.
// Mode "dotmatrix" tidak didukung karena printer dot matrix tidak
// memiliki perintah cetak bitmap raster; akan dikembalikan error 400
// agar UI dapat menampilkan pesan yang jelas kepada pengguna.
func (m *Manager) PrintImage(ctx context.Context, cfg config.Config, job ImageJob) error {
	if strings.TrimSpace(cfg.DefaultPrinter) == "" {
		return errors.New("default printer belum dikonfigurasi")
	}
	if job.Bitmap == nil {
		return errors.New("bitmap kosong")
	}

	mode := resolveMode(job.Mode, cfg.PrinterMode)
	if mode == config.PrinterModeDotMatrix {
		return errors.New("cetak gambar tidak didukung pada mode Dot Matrix / Plain Text; " +
			"pilih mode Thermal (ESC/POS) untuk mencetak gambar")
	}

	info, ok := m.Find(cfg.DefaultPrinter)
	if !ok {
		info = Info{Name: cfg.DefaultPrinter, Type: cfg.PrinterType, Status: StatusReady}
	}
	if job.Copies < 1 {
		job.Copies = 1
	}
	if job.Copies > 10 {
		job.Copies = 10
	}

	b := escpos.New().Init().Align(escpos.AlignCenter)
	if err := b.PrintRasterBitmap(job.Bitmap.Data, job.Bitmap.WidthBytes, job.Bitmap.HeightPx); err != nil {
		return fmt.Errorf("bitmap raster invalid: %w", err)
	}
	b.FeedLines(2)
	if job.CutPaper {
		b.CutPaper(true)
	}

	payload := b.Bytes()
	for i := 0; i < job.Copies; i++ {
		if err := sendToPrinter(ctx, info, payload); err != nil {
			return fmt.Errorf("gagal kirim ke printer (copy %d): %w", i+1, err)
		}
	}
	return nil
}

// alignmentByteFromString memetakan string "left|center|right" ke byte
// argumen perintah ESC a. Default = center bila tidak dikenali.
func alignmentByteFromString(pos string) byte {
	switch strings.ToLower(strings.TrimSpace(pos)) {
	case "left":
		return escpos.AlignLeft
	case "right":
		return escpos.AlignRight
	default:
		return escpos.AlignCenter
	}
}

// BuildLogoRaster melakukan decode base64, resize ke lebar kertas,
// dithering Floyd-Steinberg, dan konversi ke RasterImage.
// Helper ini dipakai oleh handler /api/print/text untuk logo opsional.
func BuildLogoRaster(b64 string, paperWidthMM int) (*RasterImage, error) {
	if strings.TrimSpace(b64) == "" {
		return nil, nil
	}
	img, _, err := imageproc.DecodeBase64Image(b64)
	if err != nil {
		return nil, err
	}
	widthDots := imageproc.MMToDots(paperWidthMM)
	mono := imageproc.PrepareMonochrome(img, widthDots, imageproc.DitherFloydSteinberg)
	data, wb, h := imageproc.ToRasterBitmap(mono)
	return &RasterImage{Data: data, WidthBytes: wb, HeightPx: h}, nil
}

// BuildImageRaster sama dengan BuildLogoRaster namun memungkinkan
// pemilihan algoritma dithering. Dipakai oleh handler /api/print/image.
func BuildImageRaster(b64 string, paperWidthMM int, dither string) (*RasterImage, error) {
	if strings.TrimSpace(b64) == "" {
		return nil, errors.New("image_base64 wajib diisi")
	}
	img, _, err := imageproc.DecodeBase64Image(b64)
	if err != nil {
		return nil, err
	}
	widthDots := imageproc.MMToDots(paperWidthMM)
	mono := imageproc.PrepareMonochrome(img, widthDots, dither)
	data, wb, h := imageproc.ToRasterBitmap(mono)
	return &RasterImage{Data: data, WidthBytes: wb, HeightPx: h}, nil
}

// ============================================================
//   DETEKSI PRINTER (SPOOLER)
// ============================================================

// detectSpoolerPrinters mengembalikan daftar printer pada OS spooler.
// Implementasi dispatch berdasarkan runtime.GOOS.
func detectSpoolerPrinters(ctx context.Context) ([]Info, error) {
	switch runtime.GOOS {
	case "windows":
		return detectSpoolerWindows(ctx)
	case "darwin", "linux":
		return detectSpoolerUnix(ctx)
	default:
		return nil, fmt.Errorf("OS %s belum didukung", runtime.GOOS)
	}
}

// detectSpoolerWindows menjalankan PowerShell Get-Printer untuk mendapatkan
// daftar printer Windows beserta status-nya. Jatuh balik ke wmic bila
// PowerShell tidak tersedia atau melebihi timeout.
func detectSpoolerWindows(ctx context.Context) ([]Info, error) {
	// Gunakan PowerShell karena lebih robust dan terstruktur.
	// Timeout 5s sudah cukup; PowerShell biasanya selesai dalam 1-3s.
	psScript := `Get-Printer | Select-Object Name,PrinterStatus | ForEach-Object { "$($_.Name)|$($_.PrinterStatus)" }`
	out, err := runCommand(ctx, 5*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", psScript)
	if err == nil {
		return parseWindowsPrinterList(out), nil
	}

	// Fallback: wmic printer (lebih ringan, tanpa PowerShell runtime).
	out2, err2 := runCommand(ctx, 5*time.Second, "wmic", "printer", "get", "name,printerstatus", "/format:csv")
	if err2 != nil {
		return nil, fmt.Errorf("powershell gagal (%v) dan wmic gagal (%v)", err, err2)
	}
	return parseWmicPrinterList(out2), nil
}

// parseWindowsPrinterList mem-parse output PowerShell "Name|PrinterStatus".
func parseWindowsPrinterList(s string) []Info {
	var out []Info
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 1 || strings.TrimSpace(parts[0]) == "" {
			continue
		}
		status := StatusReady
		if len(parts) == 2 {
			s := strings.ToLower(strings.TrimSpace(parts[1]))
			if s != "" && s != "normal" && s != "0" && s != "idle" {
				// Status non-normal: tetap "ready" kecuali jelas-jelas offline.
				if strings.Contains(s, "offline") || strings.Contains(s, "error") {
					status = StatusOffline
				}
			}
		}
		out = append(out, Info{
			Name:   strings.TrimSpace(parts[0]),
			Type:   TypeSpooler,
			Status: status,
		})
	}
	return out
}

// parseWmicPrinterList mem-parse output CSV dari wmic.
// Format: Node,Name,PrinterStatus
func parseWmicPrinterList(s string) []Info {
	var out []Info
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "node,") {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSpace(fields[1])
		if name == "" {
			continue
		}
		status := StatusReady
		if len(fields) >= 3 {
			s := strings.ToLower(strings.TrimSpace(fields[2]))
			if strings.Contains(s, "offline") || strings.Contains(s, "error") {
				status = StatusOffline
			}
		}
		out = append(out, Info{Name: name, Type: TypeSpooler, Status: status})
	}
	return out
}

// detectSpoolerUnix menggunakan lpstat -p untuk mendapatkan daftar printer
// di Linux/macOS. Format umum: "printer NAMA is idle. enabled since ...".
func detectSpoolerUnix(ctx context.Context) ([]Info, error) {
	out, err := runCommand(ctx, 5*time.Second, "lpstat", "-p")
	if err != nil {
		// lpstat mungkin tidak terinstall; bukan error fatal, kembalikan list kosong.
		return []Info{}, nil
	}
	var printers []Info
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "printer ") {
			continue
		}
		// "printer NAMA is ..."
		rest := strings.TrimPrefix(line, "printer ")
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) < 1 {
			continue
		}
		name := parts[0]
		status := StatusReady
		if strings.Contains(strings.ToLower(line), "disabled") || strings.Contains(strings.ToLower(line), "offline") {
			status = StatusOffline
		}
		printers = append(printers, Info{Name: name, Type: TypeSpooler, Status: status})
	}
	return printers, nil
}

// ============================================================
//   DETEKSI PRINTER (BLUETOOTH)
// ============================================================

// detectBluetoothPrinters dispatch berdasarkan OS. Kegagalan deteksi
// dianggap non-fatal dan menghasilkan list kosong.
func detectBluetoothPrinters(ctx context.Context) ([]Info, error) {
	switch runtime.GOOS {
	case "windows":
		return detectBluetoothWindows(ctx)
	case "linux":
		return detectBluetoothLinux(ctx)
	case "darwin":
		return detectBluetoothMac(ctx)
	default:
		return []Info{}, nil
	}
}

// detectBluetoothWindows menjalankan PowerShell Get-PnpDevice untuk
// menemukan device Bluetooth yang berasosiasi sebagai printer atau
// memiliki COM port virtual. Hasilnya difilter berdasarkan keyword.
func detectBluetoothWindows(ctx context.Context) ([]Info, error) {
	psScript := `Get-PnpDevice -Class Bluetooth -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'OK' } | ForEach-Object { $_.FriendlyName }`
	out, err := runCommand(ctx, 5*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", psScript)
	if err != nil {
		return []Info{}, nil
	}
	var infos []Info
	for _, line := range strings.Split(out, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		// Heuristik sederhana: hanya tampilkan device yang kemungkinan printer.
		l := strings.ToLower(name)
		if strings.Contains(l, "printer") || strings.Contains(l, "thermal") ||
			strings.Contains(l, "pos") || strings.Contains(l, "bt-") || strings.Contains(l, "rpp") {
			infos = append(infos, Info{Name: name, Type: TypeBluetooth, Status: StatusReady})
		}
	}
	return infos, nil
}

// detectBluetoothLinux menggunakan bluetoothctl devices untuk mendapatkan
// daftar device Bluetooth yang sudah dipasangkan.
func detectBluetoothLinux(ctx context.Context) ([]Info, error) {
	out, err := runCommand(ctx, 5*time.Second, "bluetoothctl", "devices")
	if err != nil {
		return []Info{}, nil
	}
	var infos []Info
	for _, line := range strings.Split(out, "\n") {
		// Format: "Device XX:XX:XX:XX:XX:XX FriendlyName"
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Device ") {
			continue
		}
		fields := strings.SplitN(line, " ", 3)
		if len(fields) < 3 {
			continue
		}
		name := fields[2]
		l := strings.ToLower(name)
		if strings.Contains(l, "printer") || strings.Contains(l, "thermal") ||
			strings.Contains(l, "pos") || strings.Contains(l, "bt-") || strings.Contains(l, "rpp") {
			infos = append(infos, Info{Name: name, Type: TypeBluetooth, Status: StatusReady})
		}
	}
	return infos, nil
}

// detectBluetoothMac menggunakan system_profiler SPBluetoothDataType untuk
// menemukan device Bluetooth yang terhubung pada macOS.
func detectBluetoothMac(ctx context.Context) ([]Info, error) {
	out, err := runCommand(ctx, 8*time.Second, "system_profiler", "SPBluetoothDataType")
	if err != nil {
		return []Info{}, nil
	}
	var infos []Info
	// Parser sederhana: cari baris yang berisi nama device & "Connected: Yes"
	// Banyak versi macOS punya format berbeda; kita ambil pendekatan heuristik.
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasSuffix(t, ":") && !strings.Contains(t, " ") {
			continue
		}
		l := strings.ToLower(t)
		if (strings.Contains(l, "printer") || strings.Contains(l, "thermal") || strings.Contains(l, "pos")) &&
			strings.HasSuffix(t, ":") {
			name := strings.TrimSuffix(t, ":")
			status := StatusReady
			// Cek beberapa baris setelahnya untuk status "Connected".
			for j := i + 1; j < i+8 && j < len(lines); j++ {
				s := strings.ToLower(lines[j])
				if strings.Contains(s, "connected: no") {
					status = StatusOffline
					break
				}
			}
			infos = append(infos, Info{Name: name, Type: TypeBluetooth, Status: status})
		}
	}
	return infos, nil
}

// ============================================================
//   PENGIRIMAN JOB KE PRINTER
// ============================================================

// sendToPrinter mengirim payload byte ke printer tujuan dengan retry
// otomatis (3x, jeda 5 detik) untuk transport bluetooth.
func sendToPrinter(ctx context.Context, info Info, payload []byte) error {
	switch info.Type {
	case TypeBluetooth:
		return sendBluetoothWithRetry(ctx, info, payload)
	case TypeSpooler, "":
		return sendSpooler(ctx, info, payload)
	default:
		return fmt.Errorf("tipe printer tidak dikenal: %s", info.Type)
	}
}

// sendBluetoothWithRetry membungkus sendBluetooth dengan retry 3x
// dan timeout 5 detik per percobaan. Pemilihan port:
//   - Windows: COM port pada info.Port (atau heuristik default COM3..COM9)
//   - Linux:   /dev/rfcommX pada info.Port (default /dev/rfcomm0)
func sendBluetoothWithRetry(ctx context.Context, info Info, payload []byte) error {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := sendBluetooth(attemptCtx, info, payload)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		// Jangan retry bila context utama sudah cancel.
		if ctx.Err() != nil {
			break
		}
		if attempt < 3 {
			select {
			case <-time.After(1 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("bluetooth gagal setelah 3x percobaan: %w", lastErr)
}

// sendBluetooth menulis payload ke device file/COM port yang
// merepresentasikan koneksi RFCOMM ke printer Bluetooth.
func sendBluetooth(ctx context.Context, info Info, payload []byte) error {
	port := info.Port
	if port == "" {
		// Heuristik default per OS bila port tidak diketahui.
		switch runtime.GOOS {
		case "windows":
			port = "COM3"
		case "linux":
			port = "/dev/rfcomm0"
		default:
			return fmt.Errorf("port bluetooth tidak ditentukan untuk OS %s", runtime.GOOS)
		}
	}
	// Pada Windows, COM port juga merupakan file device; bisa dibuka via os.OpenFile.
	f, err := os.OpenFile(port, os.O_WRONLY|os.O_SYNC, 0)
	if err != nil {
		return fmt.Errorf("gagal buka port %s: %w", port, err)
	}
	defer f.Close()

	// Kirim dengan goroutine + context untuk menerapkan timeout.
	doneCh := make(chan error, 1)
	go func() {
		_, werr := f.Write(payload)
		doneCh <- werr
	}()
	select {
	case werr := <-doneCh:
		return werr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// sendSpooler mengirim payload ke spooler OS dalam mode RAW.
//
// Windows: panggil WinSpool API langsung (OpenPrinterW + StartDocPrinter +
// WritePrinter + EndDocPrinter) via package golang.org/x/sys/windows.
// Pendekatan ini menghindari startup PowerShell dan Add-Type (.NET C#)
// yang sebelumnya memakan ratusan ms s/d >1 detik per cetak. Payload
// dikirim langsung dari memory tanpa file sementara.
//
// Jika WinSpool gagal (mis. printer virtual yang menolak datatype RAW,
// permission khusus, dll.) disediakan fallback klasik: tulis payload
// ke file temp lalu jalankan `print /D:` dan terakhir `copy /B` ke
// share UNC printer.
//
// Linux/macOS: tetap memakai `lp -d <printer> -o raw <tempfile>`.
func sendSpooler(ctx context.Context, info Info, payload []byte) error {
	if runtime.GOOS == "windows" {
		// Jalur cepat: WinSpool langsung tanpa file temp. Sukses di sini
		// akan menjadi kasus mayoritas dan menghilangkan overhead proses
		// eksternal sepenuhnya.
		if err := sendWinspoolRaw(ctx, info.Name, payload); err == nil {
			return nil
		} else {
			// Fallback: tulis ke file temp lalu coba `print /D:` dan
			// `copy /B` seperti jalur lama. File temp hanya dibuat
			// di jalur fallback agar jalur cepat tetap bebas I/O disk.
			return spoolWindowsFallback(ctx, info.Name, payload, err)
		}
	}

	// Non-Windows: masih butuh file temp untuk lp/lpr.
	tmp, err := os.CreateTemp("", "printbridge-*.bin")
	if err != nil {
		return fmt.Errorf("gagal membuat file temp: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("gagal menulis file temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("gagal menutup file temp: %w", err)
	}
	defer os.Remove(tmpPath)

	switch runtime.GOOS {
	case "linux", "darwin":
		return spoolUnix(ctx, info.Name, tmpPath)
	default:
		return fmt.Errorf("OS %s belum didukung untuk spooler", runtime.GOOS)
	}
}

// spoolWindowsFallback dipanggil hanya bila jalur cepat WinSpool gagal.
// Fungsi ini menulis payload ke file sementara lalu mencoba dua strategi
// legacy secara berurutan: `print /D:` dan `copy /B` ke share UNC
// printer. Error dari jalur cepat (winspoolErr) disertakan pada pesan
// akhir agar log /api/logs menampilkan konteks lengkap.
func spoolWindowsFallback(ctx context.Context, printerName string, payload []byte, winspoolErr error) error {
	tmp, err := os.CreateTemp("", "printbridge-*.bin")
	if err != nil {
		return fmt.Errorf("winspool: %v ; gagal membuat file temp: %w", winspoolErr, err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("winspool: %v ; gagal menulis file temp: %w", winspoolErr, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("winspool: %v ; gagal menutup file temp: %w", winspoolErr, err)
	}
	defer os.Remove(tmpPath)

	// Strategi 1: utilitas `print /D:` (Windows builtin).
	if _, err2 := runCommand(ctx, 15*time.Second, "cmd", "/C", "print", "/D:"+printerName, tmpPath); err2 == nil {
		return nil
	} else {
		// Strategi 2: `copy /B` ke share lokal printer (\\localhost\<printer>).
		share := `\\localhost\` + printerName
		if _, err3 := runCommand(ctx, 10*time.Second, "cmd", "/C", "copy", "/B", tmpPath, share); err3 == nil {
			return nil
		} else {
			return fmt.Errorf("winspool: %v ; print: %v ; copy: %v", winspoolErr, err2, err3)
		}
	}
}

// spoolUnix mengirim file RAW ke spooler CUPS pada Linux/macOS
// menggunakan perintah lp dengan opsi -o raw.
func spoolUnix(ctx context.Context, printerName, filePath string) error {
	out, err := runCommand(ctx, 15*time.Second, "lp", "-d", printerName, "-o", "raw", filePath)
	if err != nil {
		// Coba lpr sebagai fallback.
		_, err2 := runCommand(ctx, 15*time.Second, "lpr", "-P", printerName, "-l", filePath)
		if err2 != nil {
			return fmt.Errorf("lp: %v ; lpr: %v ; output: %s", err, err2, strings.TrimSpace(out))
		}
	}
	return nil
}

// ============================================================
//   UTIL EKSEKUSI PERINTAH
// ============================================================

// runCommand mengeksekusi sebuah perintah eksternal dengan timeout
// dan mengembalikan stdout-nya sebagai string. Stdin di-set kosong.
func runCommand(parent context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader("")
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%s gagal: %w (stderr=%s)", filepath.Base(name), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// drainReaderToString sederhana, dipakai bila ingin membaca pipa output
// secara streaming. Disimpan untuk kelengkapan API.
func drainReaderToString(r io.Reader) string {
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
