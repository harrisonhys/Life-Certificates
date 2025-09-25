📄 Product Requirement Document (PRD)

Project: Life Certificate Service (LCS)
Owner: Harianto
Date: September 2025

1. Latar Belakang

Perusahaan pengelola dana pensiun wajib memastikan penerima manfaat masih hidup (proof of life) sebelum mencairkan dana.
Metode manual (hadir fisik, tanda tangan dokumen, validasi bank/pos) memiliki kelemahan:

Menyulitkan peserta lansia, terutama yang tinggal jauh.

Membutuhkan biaya operasional besar.

Masih ada celah fraud (klaim atas nama peserta wafat).

Untuk itu, dikembangkan Life Certificate Service (LCS) sebagai microservice bisnis di atas FR Core Service.

2. Tujuan

Memberikan mekanisme Life Certificate digital berbasis Face Recognition & Liveness Detection.

Memastikan pembayaran pensiun hanya ke peserta yang sah.

Menyediakan audit trail status verifikasi peserta.

Memisahkan logika bisnis (LCS) dari engine biometrik (FR Core).

3. Lingkup
In-Scope

Registrasi peserta dengan referensi ke FR Core.

Verifikasi berkala Life Certificate dengan selfie/liveness check.

Penyimpanan status verifikasi berkala (VALID/INVALID/REVIEW).

API untuk admin & peserta (cek status terakhir).

Integrasi ke FR Core untuk face recognition.

Out-of-Scope

Encoding/face matching engine (ditangani FR Core).

Sistem pembayaran pensiun.

Portal/UI frontend (akan konsumsi API LCS).

4. Use Case Detail
UC-01: Registrasi Peserta

Aktor: Admin (internal perusahaan).

Flow:

Admin input data peserta (NIK, nama, dsb.).

Selfie peserta awal diambil (KTP/selfie langsung).

LCS memanggil FR Core /upload untuk menyimpan encoding wajah.

FR Core mengembalikan fr_ref.

LCS menyimpan data peserta + fr_ref.

UC-02: Verifikasi Life Certificate

Aktor: Peserta pensiun.

Flow:

Peserta login ke aplikasi → ambil selfie (atau video singkat).

LCS menerima file.

Jalankan liveness detection (contoh: kedipan, gerakan kepala).

Jika PASS → kirim ke FR Core /recognize (atau /recognize-multi).

FR Core mengembalikan hasil: label, similarity, distance.

LCS cek apakah distance <= threshold (0.6) dan label cocok dengan participant.fr_ref.

Jika ya → status = VALID, else → INVALID.

Jika liveness FAIL → status = REVIEW.

Simpan ke tabel life_certificate.

UC-03: Cek Status

Aktor: Peserta / Admin.

Flow:

Request GET /life-certificate/status/{participant_id}.

LCS ambil status terakhir.

Return status (VALID/INVALID/REVIEW + timestamp).

5. Arsitektur Sistem
flowchart TD
    P[Peserta] -->|Selfie| LCS[Life Certificate Service]
    A[Admin] -->|Registrasi Peserta| LCS
    LCS -->|Upload/Recognize| FR[FR Core Service]
    FR --> FRDB[(FR Database: encoding, image)]
    LCS --> LCDB[(LCS Database: peserta, status verifikasi)]

6. Database Design
Table: participants
Kolom	Tipe	Deskripsi
id	UUID	Primary key
nik	VARCHAR(20)	Nomor Induk Kependudukan
name	VARCHAR(100)	Nama peserta
fr_ref	UUID	Referensi ke FR Core
created_at	TIMESTAMP	Tanggal registrasi
updated_at	TIMESTAMP	Update terakhir
Table: life_certificate
Kolom	Tipe	Deskripsi
id	UUID	Primary key
participant_id	UUID	Relasi ke participants
selfie_path	TEXT	Lokasi penyimpanan selfie
status	ENUM	VALID / INVALID / REVIEW
distance	FLOAT	Jarak hasil recognition
similarity	FLOAT	Nilai similarity
verified_at	TIMESTAMP	Timestamp verifikasi
notes	TEXT	Catatan tambahan/manual review
7. API Specification
POST /participants/register

Request:

{
  "nik": "1234567890123456",
  "name": "Harianto",
  "image": "base64/selfie.jpg"
}


Response:

{
  "status": "success",
  "data": {
    "participant_id": "uuid",
    "fr_ref": "uuid-fr-core"
  }
}

POST /life-certificate/verify

Request (multipart/form-data):

POST /life-certificate/verify
Content-Type: multipart/form-data

participant_id=uuid
image=@selfie.jpg


Response (VALID):

{
  "status": "success",
  "data": {
    "participant_id": "uuid",
    "verification_status": "VALID",
    "similarity": 92.5,
    "verified_at": "2025-09-25T10:30:00Z"
  }
}


Response (INVALID):

{
  "status": "success",
  "data": {
    "participant_id": "uuid",
    "verification_status": "INVALID",
    "similarity": 40.2,
    "verified_at": "2025-09-25T10:31:00Z"
  }
}

Response (REVIEW – gagal liveness):

{
  "status": "success",
  "data": {
    "participant_id": "uuid",
    "verification_status": "REVIEW",
    "verified_at": "2025-09-25T10:32:00Z"
  }
}

GET /life-certificate/status/{participant_id}

Response:

{
  "status": "success",
  "data": {
    "participant_id": "uuid",
    "last_status": "VALID",
    "similarity": 93.1,
    "verified_at": "2025-09-25T10:30:00Z"
  }
}

8. Flow Detail
Sequence Diagram UC-02: Verifikasi
sequenceDiagram
    participant P as Peserta
    participant LCS as Life Certificate Service
    participant FR as FR Core Service
    participant FRDB as FR DB
    participant LCDB as LC DB

    P->>LCS: POST /life-certificate/verify (selfie)
    LCS->>LCS: Jalankan Liveness Detection
    alt Liveness PASS
        LCS->>FR: POST /recognize (selfie)
        FR->>FRDB: Cari encoding terdekat
        FRDB-->>FR: Hasil encoding match
        FR-->>LCS: Label + similarity + distance
        alt Distance <= Threshold (0.6)
            LCS->>LCDB: Simpan status VALID
            LCS-->>P: Response VALID
        else Distance > Threshold
            LCS->>LCDB: Simpan status INVALID
            LCS-->>P: Response INVALID
        end
    else Liveness FAIL
        LCS->>LCDB: Simpan status REVIEW
        LCS-->>P: Response REVIEW
    end

9. Non-Functional Requirements

Performance: verifikasi ≤ 2 detik.

Scalability: mampu handle ≥ 100k peserta aktif.

Security: data biometrik hanya di FR Core, LCS hanya simpan referensi.

Compliance: patuh UU PDP (perlindungan data pribadi).

Availability: uptime ≥ 99%.

10. Risiko & Mitigasi
Risiko	Dampak	Mitigasi
Peserta lansia kesulitan selfie	Verifikasi gagal	UI sederhana, bantuan keluarga
FR Core down	LCS tidak bisa verifikasi	Retry + fallback manual review
Liveness gagal karena kamera jelek	False REVIEW	Opsi retry
Encoding wajah berubah karena usia	False INVALID	Update encoding berkala
11. Future Enhancement

Reminder otomatis: notifikasi WA/SMS/email untuk peserta yang harus verifikasi.

Geo-tagging: lokasi selfie untuk audit tambahan.

Video-based liveness: lebih aman daripada foto statis.

Integrasi offline: bank/pos bisa jadi fallback channel.


Service face recognition core ada di postman collection berikut yang ada di root project

