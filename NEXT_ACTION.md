# NEXT ACTION

Updated: 2026-09-19

## Published candidate

**v1.3.2-test.6** is published.

Automated verification:
- PR State Guard `35424935327` — PASS
- PR Windows Portable `35424935422` — PASS
- Main State Guard `35424975533` — PASS
- Main Windows Portable `35424975537` — PASS
- Publish Owner Test Candidate `35424975501` — PASS
- unit tests / vet / Windows x64 cross-build / public binary scan / packaging — PASS

## What changed

Normal operation no longer requires a separate runtime-profile file.

### Thiết lập now has two sources

1. **Sản lượng**
2. **Đang lấy hàng**

For each source:
1. select the source;
2. paste the matching Dashboard cURL (bash);
3. click **LƯU NGUỒN**;
4. click **KIỂM TRA NGUỒN**.

The app stores:
- request URL/method/body/non-sensitive headers;
- current Dashboard session separately;
- all local operational configuration protected with Windows DPAPI.

The raw cURL is cleared from the input and private endpoint/session data never enters the public repository/release.

## Daily date handling

Dates found in the captured request are converted to previous/current/next-day templates and expanded on each sync, preventing a cURL captured today from staying pinned to today's literal date.

## Sync pipeline

**Sản lượng**
→ Payroll/Productivity XLSX
→ User/PDA
→ Phân ca
→ Pick
→ Pack.

**Đang lấy hàng**
→ Active Picking JSON
→ live status/progress table.

The two sources are independent. If one source fails or is not yet configured, successful data continues to process and previous valid snapshots are retained.

## Owner test now

1. Update/run **test.6**.
2. Thiết lập → chọn **Sản lượng** → paste matching cURL → **LƯU NGUỒN**.
3. **KIỂM TRA NGUỒN**.
4. Chọn **Đang lấy hàng** → paste matching cURL → **LƯU NGUỒN**.
5. **KIỂM TRA NGUỒN**.
6. Press **ĐỒNG BỘ**.
7. Verify:
   - User/PDA has data;
   - Phân ca has data;
   - Pick has expected rows/calculations;
   - Pack has expected rows/calculations;
   - Đang lấy hàng has current live rows.

If a source test returns OK but resulting tables/calculations are incorrect, export/send the test.6 Log. That becomes the next parser/business-parity fix.
