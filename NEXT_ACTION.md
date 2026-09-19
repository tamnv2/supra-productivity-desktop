# NEXT ACTION

Updated: 2026-09-19

## Published candidate

**v1.3.2-test.7** is published.

Automated verification:
- PR State Guard `35426884976` — PASS
- PR Windows Portable `35426884987` — PASS
- Main State Guard `35426934084` — PASS
- Main Windows Portable `35426934063` — PASS
- Publish Owner Test Candidate `35426934057` — PASS
- unit tests / vet / Windows x64 cross-build / public binary scan / packaging — PASS

## Corrected architecture

Owner configures **one Dashboard cURL/session only**.

The app derives two internal requests from that same protected session:
- current-picking JSON for **Đang lấy hàng**;
- payroll/productivity XLSX export for the Excel production calculation path.

Existing test.6 local credentials are migrated automatically on startup where the saved cURL is still available.

## Excel-style processing

Payroll/Productivity XLSX
→ normalize/map production fields
→ enrich missing employee fields from current-picking data
→ aggregate User + Job
→ Phân ca auto/manual
→ classify in-shift/overtime and even/odd
→ Pick
→ Pack.

User/PDA combines production users with live employee/PDA/client information.

Additional regression fixes:
- Site filter is 1291, not reconstructed 1921.
- Sync test validates both response parsers, not only HTTP 200.
- Previous valid snapshots remain if either internal request fails.
- Log records sanitized row counts for payroll, User/PDA, Pick, Pack, Phân ca and active rows.

## Owner test now

1. Update/run test.7.
2. Open Thiết lập: it should show Dashboard already configured after migration.
3. Press **KIỂM TRA ĐỒNG BỘ**.
4. Press **ĐỒNG BỘ**.
5. Verify non-empty/expected:
   - User / PDA
   - Đang lấy hàng
   - Pick
   - Pack
   - Phân ca
6. Compare several users/DO/pieces/shift rows directly with the Excel workbook.

If anything differs, export/send the test.7 Log. Do not re-enter the same cURL twice.
