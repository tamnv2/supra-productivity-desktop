# LOCAL DASHBOARD CONFIGURATION

Updated: 2026-09-19

The repository is public. Production host names, credentials, tokens, raw cURL values and company data must never be committed.

## Owner workflow

The app uses **one Dashboard cURL/session**.

Owner setup:
1. open Thiết lập;
2. paste one valid Dashboard cURL (bash);
3. click **LƯU CẤU HÌNH**;
4. click **KIỂM TRA ĐỒNG BỘ**.

The application derives the required internal requests from the captured Dashboard origin and uses the same protected session for both:
- current **Đang lấy hàng** data;
- **Sản lượng** export for the Excel-style productivity pipeline.

The full private origin and credentials are stored only under the current Windows user with DPAPI. They are not embedded in the public source or release.

## Runtime pipeline

### Đang lấy hàng
Dashboard session
→ current-picking JSON
→ 31-column live table
→ status/warning/progress/device fields.

### Sản lượng
Same Dashboard session
→ payroll/productivity XLSX export for yesterday..today
→ parse completed production rows
→ enrich employee attributes from current-picking data
→ Mapping-equivalent normalization
→ Phân ca
→ classify in-shift/overtime and even/odd
→ Pick / Pack.

### User / PDA
The native app joins production users with current-picking employee/device fields and deduplicates by User. This provides the employee reference needed by the production calculation without requiring Owner to configure a second Dashboard source.

## Compatibility

test.6 stored two source bindings when the same cURL was entered twice. On startup, the next version automatically rebuilds both internal bindings from the already saved single Dashboard cURL/session. Owner should not need to enter the same cURL twice again.

## Failure behavior

- If current-picking succeeds and payroll fails, the live table remains updated and previous valid production tables remain.
- If payroll succeeds and current-picking fails, production is still processed and previous valid current-picking data remains.
- A failed request never clears the previous valid snapshot.
- Logs record request type, HTTP status, bytes, parse row count and processing row counts without endpoint or credential values.
