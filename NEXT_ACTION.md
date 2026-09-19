# NEXT ACTION

Updated: 2026-09-19

## Published candidate

**v1.3.2-test.8** is published and all automated gates passed.

Automated verification:
- PR Project State Guard `35440770198` — PASS
- PR Windows Portable `35440770194` — PASS
- Main Project State Guard `35442973560` — PASS
- Main Windows Portable `35442973548` — PASS
- Publish Owner Test Candidate `35442973570` — PASS
- unit tests / vet / Windows x64 cross-build / public binary scan / packaging — PASS
- release assets present: `SupraProductivity.exe`, `SHA256SUMS.txt`, `SupraProductivity_v1.3.2-test.8_windows_x64.zip`

## test.8 behavior to verify on target laptop

1. Launch test.8: **Từ ngày / Đến ngày default to today**.
2. Business data is refreshed only by **ĐỒNG BỘ**; no production minute-based auto-sync runs.
3. In **BÁO CÁO**, verify four native views:
   - Recap
   - % chẵn lẻ
   - NSLD Pick
   - NSLD Pack
4. Select a historical range with partial cache, e.g. 01/09→19/09 where 01/09→15/09 already exists:
   - historical cached days must not be re-downloaded blindly;
   - only missing contiguous ranges are requested;
   - today is refreshed whenever manual sync is pressed.
5. Compare report totals/rates with the reference workbook/Google Sheet logic.
6. Verify **Pick / Pack / Phân ca use only the selected Đến ngày**, while Báo cáo uses the full selected range.
7. Verify Pick Auto PP matches the workbook first-match `Mã Tham chiếu → Pack user` behavior.

If any live value differs, export the test.8 Log for targeted correction. Do not re-enter the Dashboard cURL unless the saved session itself is invalid.
