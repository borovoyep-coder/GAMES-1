# Каменный рубеж для Windows

Один переносимый файл `StoneLine-Windows-x64.exe`, без установщика и прав администратора.
Игра v19 встроена в EXE; после установки среды WebView2 интернет для самой игры не нужен.

Требования: Windows 10/11 x64 и Microsoft Edge WebView2 Runtime. Среда обычно уже
установлена. Если ее нет, приложение предложит открыть официальную страницу Microsoft.

Прогресс хранится в `%LOCALAPPDATA%\StoneLine\WebView2`. Переименование или перенос
EXE не меняют папку профиля. Удаление EXE не стирает сохранение.
Встроенная игра сохраняет текущий бой примерно раз в две секунды.
Сохранения веб-версии и приложения отдельные. Для переноса в старой игре откройте
щит уровня аккаунта, затем «Профиль и сохранения» → «Экспорт прогресса».
Импортируйте полученный JSON через то же окно в Windows-приложении.

## Сборка

Нужен Go 1.27.4+. Исходный HTML и Windows-ресурсы включены в этот каталог.

```
go mod download
python3 build.py
```

В Windows с Python Launcher можно использовать `py build.py`.
Альтернатива без Python в PowerShell:

```
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -trimpath -buildvcs=false '-ldflags=-s -w -H=windowsgui -buildid=' -o StoneLine-Windows-x64.exe .
```

Базовые проверки подготовки файлов и сохранения существующего профиля: `go test .`.
В этой поставке Windows-EXE собран перекрестной компиляцией в Linux. Нативный запуск
в Windows не проверялся. Цифровая подпись разработчика отсутствует.

`assets/THIRD-PARTY-NOTICES.txt` содержит лицензии среды Go, загрузчика WebView2
и используемых библиотек. При запуске копия сохраняется рядом с профилем приложения.

Перегенерация Windows-ресурсов (не нужна для обычной сборки):

```
go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico assets/app.ico -manifest app.manifest -o rsrc_windows_amd64.syso
```
