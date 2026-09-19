# NEXT ACTION

Updated: 2026-09-19

## Current finding

**v1.3.2-test.5** fixed the repeated GUI lag/Not Responding reported by the Owner.

The remaining sync failure is independent:
- Dashboard credentials are present;
- no live source request definitions are stored;
- therefore test.5 reports the internal `SYNC_PROFILE_MISSING` condition.

This is an application integration gap, not an Owner error.

## test.6 candidate

Normal Owner operation no longer uses a separate runtime-profile file.

### Thiết lập

Two source choices:
1. **Sản lượng**
2. **Đang lấy hàng**

One-time setup per source:
1. select source;
2. paste matching Dashboard cURL (bash);
3. click **LƯU NGUỒN**;
4. click **KIỂM TRA NGUỒN**.

The app:
- extracts URL/method/body/non-sensitive headers;
- merges current Dashboard session values;
- stores source + session with Windows DPAPI;
- clears raw cURL from the input;
- never commits the private request to public GitHub.

### Daily date handling

Captured previous/current/next dates are templated and expanded at sync time so tomorrow's sync does not reuse today's literal date.

### Đồng bộ pipeline

**Sản lượng**
→ XLSX
→ Payroll/Productivity parser
→ User/PDA
→ Phân ca
→ Pick/Pack.

**Đang lấy hàng**
→ JSON
→ Active Picking parser
→ live progress/status table.

The two sources run independently. A failure/missing source does not clear the last valid data of the other source.

## Immediate next action

Run PR CI for **v1.3.2-test.6**.

After publish, Owner test:
1. configure **Sản lượng** once;
2. test it;
3. configure **Đang lấy hàng** once;
4. test it;
5. press **ĐỒNG BỘ**;
6. verify populated User/PDA, Phân ca, Pick, Pack and Đang lấy hàng;
7. send exported Log if a source request succeeds but parsing/business output is wrong.
