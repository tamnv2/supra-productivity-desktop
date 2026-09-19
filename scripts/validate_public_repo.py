import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

SKIP_DIRS = {".git", ".github/workflows/__pycache__", "bin", "build", "dist", "out", "tmp"}
TEXT_EXTS = {".md", ".json", ".jsonl", ".py", ".go", ".mod", ".sum", ".yml", ".yaml", ".txt", ".ps1", ".cmd"}

patterns = [
    ("authorization/token value", re.compile(r"(?i)(authorization|access[_ -]?token|bearer)[ 	]*[:=][ 	]*['\"]?(?:bearer[ 	]+)?[A-Za-z0-9._~+\-/=]{24,}")),
    ("APISID/USID quoted literal", re.compile(r"(?i)\b(APISID|USID)[ 	]*[:=][ 	]*['\"][^'\"\n]{12,}['\"]")),
    ("signature value", re.compile(r"(?i)\bx-signature(?:-nonce)?[ 	]*[:=][ 	]*['\"]?[A-Za-z0-9._~+\-/:=]{12,}")),
    ("cookie value", re.compile(r"(?i)\bcookie[ 	]*[:=][ 	]*['\"][^'\"\n]{30,}")),
    ("private key", re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----")),
]

violations = []
for path in ROOT.rglob("*"):
    if not path.is_file():
        continue
    rel = path.relative_to(ROOT)
    if any(part in SKIP_DIRS for part in rel.parts):
        continue
    if path.suffix.lower() not in TEXT_EXTS and path.name not in {".gitignore"}:
        continue
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        continue
    for label, rx in patterns:
        if rx.search(text):
            violations.append(f"{rel}: suspected {label}")

if violations:
    raise SystemExit("Public-repo guard failed:\n" + "\n".join(violations))

print("public repo sensitive-value scan: OK")
