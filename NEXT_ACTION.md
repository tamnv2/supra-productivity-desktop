# NEXT ACTION

Updated: 2026-09-19

## Candidate being validated

**v1.3.2-test.8** — native Recap + smart date cache.

Implementation on `feat/native-recap-smart-date-cache`:
- report range defaults to today and accepts Từ ngày / Đến ngày;
- no production auto-sync interval; business refresh is manual through **ĐỒNG BỘ**;
- native **BÁO CÁO** tab: Recap, % chẵn lẻ, NSLD Pick, NSLD Pack;
- per-day local payroll cache;
- historical sync downloads only missing contiguous ranges;
- today always refreshes on manual sync;
- reports use full selected range; Pick/Pack/Phân ca use selected end date;
- Excel first-match Mã Tham chiếu → Pack user rule retained for Pick Auto PP.

## Next automated gate

1. Open PR to `main`.
2. Require PR State Guard and Windows Portable workflow to PASS:
   - project/public-repo guards;
   - `go test ./...`;
   - `go vet ./...`;
   - Windows x64 build and binary public scan.
3. Merge only after CI passes.
4. Merge changes `release/test-version.txt` to `v1.3.2-test.8`, which triggers owner-test prerelease publication.
5. Verify main build + publish run before handing the EXE to Owner.

## Owner runtime test after publication

1. Launch test.8: date range must default to today's date.
2. In **BÁO CÁO**, select a historical range with partial local cache and press **ĐỒNG BỘ**.
3. Log/status must show only missing days downloaded; already-cached historical days are not requested again.
4. Verify four report views against the reference workbook:
   - Recap
   - % chẵn lẻ
   - NSLD Pick
   - NSLD Pack
5. Verify Pick/Pack/Phân ca show only the selected **Đến ngày**, not a multi-day aggregation.
6. Repeat manual sync on today: today's cache must refresh while prior historical days remain reused.
7. Export Log if any value differs; credentials/raw company data must remain absent from GitHub.
