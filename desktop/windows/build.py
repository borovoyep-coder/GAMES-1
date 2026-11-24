#!/usr/bin/env python3
"""Cross-build the portable Windows x64 application. Requires Go 1.27.4+."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parent
output = root / 'release'
output.mkdir(exist_ok=True)
env = os.environ.copy()
env.update(GOOS='windows', GOARCH='amd64', CGO_ENABLED='0')
subprocess.run([env.get('STONE_LINE_GO', 'go'), 'build', '-trimpath', '-buildvcs=false',
                '-ldflags=-s -w -H=windowsgui -buildid=', '-o',
                str(output / 'StoneLine-Windows-x64.exe'), '.'], cwd=root, env=env, check=True)
print(output / 'StoneLine-Windows-x64.exe')
