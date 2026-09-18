import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
state_path = ROOT / "ops" / "project-state.json"
next_path = ROOT / "NEXT_ACTION.md"

required = {
    "schema_version",
    "project_id",
    "project_name",
    "repository",
    "updated_at",
    "current_version",
    "current_phase",
    "current_status",
    "current_task",
    "next_action",
    "blockers",
    "known_issues",
    "test_status",
    "security",
    "canonical_docs",
}

state = json.loads(state_path.read_text(encoding="utf-8"))
missing = sorted(required - set(state))
if missing:
    raise SystemExit(f"project-state.json missing required keys: {missing}")

if state["repository"] != "tamnv2/supra-productivity-desktop":
    raise SystemExit("Unexpected repository identity in project-state.json")

if state["security"].get("public_repo") is not True:
    raise SystemExit("Public-repo security flag must remain true")

if not next_path.exists() or len(next_path.read_text(encoding="utf-8").strip()) < 40:
    raise SystemExit("NEXT_ACTION.md is missing or too small")

print("project state: OK")
