#!/usr/bin/env python3
from __future__ import annotations

import re
import sys
from pathlib import Path

ALLOWED_URL_HOSTS = {"github.com", "api.github.com", "schemas.openxmlformats.org", "www.w3.org"}
URL_RX = re.compile(rb"https?://[A-Za-z0-9._~:/?#\[\]@!$&'()*+,;=%-]+")
BEARER_RX = re.compile(rb"(?i)bearer[ \t]+[A-Za-z0-9._~+\-/=]{32,}")
PRIVATE_KEY_RX = re.compile(rb"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----")

def printable_strings(data: bytes, minimum: int = 8):
    buf = bytearray()
    for b in data:
        if 32 <= b <= 126:
            buf.append(b)
        else:
            if len(buf) >= minimum:
                yield bytes(buf)
            buf.clear()
    if len(buf) >= minimum:
        yield bytes(buf)

def host_from_url(raw: bytes) -> str:
    s = raw.decode("ascii", "ignore")
    rest = s.split("://", 1)[-1]
    hostport = rest.split("/", 1)[0].split("@")[-1]
    return hostport.split(":", 1)[0].lower()

def main() -> int:
    if len(sys.argv) != 2:
        print("usage: scan_binary.py <exe>", file=sys.stderr)
        return 2
    data = Path(sys.argv[1]).read_bytes()
    failures: list[str] = []
    if PRIVATE_KEY_RX.search(data):
        failures.append("embedded private key marker")
    for s in printable_strings(data):
        if BEARER_RX.search(s):
            failures.append("embedded Bearer-like credential")
        for raw in URL_RX.findall(s):
            host = host_from_url(raw)
            if "." not in host:
                continue
            if host not in ALLOWED_URL_HOSTS:
                failures.append(f"non-public hard-coded URL host: {host}")
    if failures:
        print("binary public-release guard failed:", file=sys.stderr)
        for x in sorted(set(failures)):
            print(f"- {x}", file=sys.stderr)
        return 1
    print("binary public-release guard: OK")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
