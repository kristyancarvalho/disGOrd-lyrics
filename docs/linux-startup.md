# Linux Startup

DisGOrd Lyrics can run as a user-level systemd service. Root access is not required.

Install the binary:

```sh
install -Dm755 disgord-lyrics ~/.local/bin/disgord-lyrics
```

Initialize and edit the configuration:

```sh
~/.local/bin/disgord-lyrics init
~/.local/bin/disgord-lyrics config-path
```

Create `~/.config/systemd/user/disgord-lyrics.service`:

```ini
[Unit]
Description=DisGOrd Lyrics
After=graphical-session.target

[Service]
ExecStart=/home/YOUR_USER/.local/bin/disgord-lyrics run
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

Replace `YOUR_USER` with the account name. systemd does not expand `~` in `ExecStart`.

Reload the user service manager and enable the service:

```sh
systemctl --user daemon-reload
systemctl --user enable --now disgord-lyrics.service
```

Inspect its state and logs:

```sh
systemctl --user status disgord-lyrics.service
journalctl --user -u disgord-lyrics.service
```

The service must run in the same user session and D-Bus environment as the MPRIS media player.

To stop and disable it:

```sh
systemctl --user disable --now disgord-lyrics.service
```
