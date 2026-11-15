# Patched gogio

This is `gioui.org/cmd/gogio` from `gioui.org/cmd` v0.10.3 (dual-licensed
Unlicense and MIT, see LICENSE), with two additions for the Android app:

- **`-androidapp <file>`:** the XML in the file goes inside the Android
  manifest's `<application>` element.
- **`-androidperms <list>`:** extra `<uses-permission>`s, comma-separated.

Bedrocktool needs these to declare its file provider and its foreground
service (`ui/gui/android/manifest-application.xml`) and the service's
permissions. The provider gives Minecraft a `content://…/World.mcworld` link
it can import, and the service keeps the app running in the background.
Stock gogio has no way to add either.

The changes are marked `bedrocktool` in `gogio/main.go` and
`gogio/androidbuild.go`. Gogio's tests and test data were left out.

This directory is its own Go module, outside the repository's `go.work`.
`build.go` builds it to `./gogio` for Android builds. By hand:

```sh
cd gio-cmd && GOWORK=off go build -o ../gogio.exe ./gogio
```

To update it, copy `gogio/` from a newer `gioui.org/cmd` and apply the same
small changes again.
