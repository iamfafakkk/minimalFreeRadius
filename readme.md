**FreeRADIUS + MySQL Installer (Ubuntu 22.04)**

- **Tujuan:** Menginstal dan mengonfigurasi FreeRADIUS 3.0 dengan backend MySQL yang siap uji login (radtest/radclient).
- **Sistem:** Ubuntu 22.04 LTS (mendukung varian minor), koneksi internet aktif, jalankan sebagai `root` atau `sudo`.
- **Script:** `install.sh`

**Apa yang diperbaiki**

- **update_system:** Ditambahkan kembali agar aman dan ringan (tanpa full upgrade), sekaligus memasang utilitas seperti `net-tools`.
- **MySQL root:** Set password root menggunakan koneksi socket, bukan preseed, sehingga konsisten di Ubuntu 22.04.
- **Izin FreeRADIUS:** Tidak lagi mengubah permission secara rekursif yang bisa merusak layanan.
- **Konfigurasi SQL:** Mengaktifkan `sql` di file site (`default` dan `inner-tunnel`) dengan regex berbasis blok, bukan nomor baris.
- **Test users:** Idempotent (hapus dulu jika ada) agar script aman dijalankan ulang.

**Cara Menjalankan**

1) Jalankan sebagai root
   - `chmod +x install.sh`
   - `sudo ./install.sh`

2) Ikuti konfirmasi di awal proses.

3) Script akan:
   - Memasang paket dasar + MySQL + FreeRADIUS.
   - Mengatur password MySQL root dan membuat DB `radius` + user `radius`.
   - Mengaktifkan modul SQL dan impor schema.
   - Membuat akun uji (`testuser`, `admin`, `user1`).
   - Memvalidasi layanan + membuat status report di `/tmp/freeradius_status.txt`.

**Uji Autentikasi**

- `radtest testuser testpass localhost 1812 testing123`
- `echo "User-Name = 'testuser', User-Password = 'testpass'" | radclient -x localhost:1812 auth testing123`

Jika sukses, Anda akan melihat `Access-Accept`.

**Manajemen Layanan**

- Start: `systemctl start freeradius`
- Stop: `systemctl stop freeradius`
- Restart: `systemctl restart freeradius`
- Status: `systemctl status freeradius`
- Debug: `freeradius -X`

**Lokasi Konfigurasi Penting**

- Main: `/etc/freeradius/3.0/radiusd.conf`
- Modul SQL: `/etc/freeradius/3.0/mods-available/sql`
- Site default: `/etc/freeradius/3.0/sites-available/default`
- Inner-tunnel: `/etc/freeradius/3.0/sites-available/inner-tunnel`
- Clients: `/etc/freeradius/3.0/clients.conf`

**Kredensial Default (harap ganti untuk produksi)**

- MySQL root: `radius123!`
- DB user: `radius` / `radiuspass123!`
- RADIUS shared secret (lokal): `testing123`

Ubah nilai di `install.sh:27` bagian variabel global sesuai kebutuhan.

**Catatan**

- Script divalidasi untuk Ubuntu 22.04. Untuk versi lain, Anda masih bisa lanjut, namun ada peringatan kompatibilitas.
- Jika jaringan memblokir ICMP (ping), validasi internet awal dapat gagal. Pastikan konektivitas paket APT berjalan.
- Bila `radtest` gagal, lihat log: `tail -f /var/log/freeradius/radius.log` atau jalankan `freeradius -X` untuk debug rinci.

Selesai. Jika Anda ingin saya menambahkan opsi custom (mis. `read_clients = yes` dari DB `nas`), beri tahu preferensinya.
