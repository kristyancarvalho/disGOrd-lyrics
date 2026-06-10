# Windows Startup

Windows media detection is not supported in v0.1.0. Do not enable automatic startup for this release. These instructions can be used after Windows media support is added.

## Startup Folder

1. Extract `disgord-lyrics.exe` to a stable directory such as `%LOCALAPPDATA%\DisGOrd Lyrics`.
2. Run `disgord-lyrics.exe init` and edit the generated configuration.
3. Press `Win+R`.
4. Enter `shell:startup`.
5. Create a shortcut to `disgord-lyrics.exe`.
6. Edit the shortcut target and append `run`.

The shortcut target should resemble:

```text
"C:\Users\YOUR_USER\AppData\Local\DisGOrd Lyrics\disgord-lyrics.exe" run
```

Remove the shortcut to disable startup.

## Task Scheduler

1. Open Task Scheduler.
2. Select `Create Task`.
3. Use a name such as `DisGOrd Lyrics`.
4. Add an `At log on` trigger for the current user.
5. Add a `Start a program` action.
6. Select `disgord-lyrics.exe` as the program.
7. Enter `run` in `Add arguments`.
8. Do not enable `Run with highest privileges`.
9. Save the task.

Run the task manually once to inspect its result. Administrator privileges are not required.
