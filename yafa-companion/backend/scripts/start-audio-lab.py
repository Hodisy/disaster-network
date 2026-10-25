#!/usr/bin/env python3
"""Build/run the standalone native Go bench. No Docker, memory or bus required."""
import os
import secrets
import signal
import shlex
import subprocess
from pathlib import Path

backend = Path(__file__).resolve().parents[1]
private = backend / '.env.audio-lab'
if not private.exists():
    fd = os.open(private, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, 'w') as output:
        output.write('# Private local audio lab. Never commit.\nYAFA_AUDIO_LAB_TOKEN=' + secrets.token_hex(32) + '\nOPENAI_API_KEY=\nYAFA_TRANSCRIBE_MODEL=gpt-transcribe\nYAFA_FFMPEG=\n')
os.chmod(private, 0o600)
environment = dict(os.environ)
for line in private.read_text().splitlines():
    if '=' not in line or line.lstrip().startswith('#'): continue
    name, raw = line.split('=', 1)
    if name not in ('YAFA_AUDIO_LAB_TOKEN', 'OPENAI_API_KEY', 'YAFA_TRANSCRIBE_MODEL', 'YAFA_FFMPEG', 'OPENAI_LIVE_MODEL', 'OPENAI_LIVE_VOICE', 'OPENAI_LIVE_BACKEND_MODEL'): continue
    value = shlex.split(raw, comments=True)
    if value and value[0]: environment[name] = value[0]
if len(environment.get('YAFA_AUDIO_LAB_TOKEN', '')) < 32:
    raise SystemExit('Fill YAFA_AUDIO_LAB_TOKEN with a private random value of at least 32 characters')
environment.setdefault('GOCACHE', '/private/tmp/yafa-go-cache')
(backend / 'bin').mkdir(exist_ok=True)
binary = backend / 'bin' / 'yafa-audio-lab'
subprocess.run(['go', 'build', '-o', str(binary), './cmd/audio-lab'], cwd=backend, env=environment, check=True)
print('Starting the standalone audio lab. Provider keys and recordings are not logged.', flush=True)
process = subprocess.Popen([str(binary)], cwd=backend, env=environment)
def interrupted(*_):
    raise KeyboardInterrupt()
signal.signal(signal.SIGTERM, interrupted)
try:
    raise SystemExit(process.wait())
except KeyboardInterrupt:
    process.terminate()
    try: process.wait(timeout=22)
    except subprocess.TimeoutExpired:
        process.kill(); process.wait()
