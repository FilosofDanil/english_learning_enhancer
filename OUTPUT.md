# OUTPUT — сборка и запуск

## Шаги (Windows / macOS / Linux из корня `english_learn_tg_bot`)

1. Установите [Go ≥ 1.22](https://go.dev/dl/) (в проекте указан `go 1.24.3`; подойдёт совместимая версия вашего toolchain).

2. Скопируйте `.env.example` в `.env` и задайте токен:

   ```powershell
   copy .env.example .env   # или cp на Unix
   ```

   Заполните `TELEGRAM_BOT_TOKEN=` токеном от [@BotFather](https://t.me/BotFather).

3. При необходимости поменяйте `CONTENT_PATH` (по умолчанию `CONTENT.md` в корне проекта).

4. Подтягивание модулей (если нужно заново):

   ```powershell
   go mod download
   ```

5. Тесты:

   ```powershell
   go test ./...
   ```

6. Запуск бота из корня (чтобы `CONTENT.md` резолвился по умолчанию):

   ```powershell
   go run .
   ```

   Сборка бинарника:

   ```powershell
   go build -o telegram-quiz-bot.exe .
   .\telegram-quiz-bot.exe
   ```

7. Команды в Telegram: `/start`, `/test`, `/print`, `/stop` (см. [DOC.md](DOC.md)).

## Устранение ошибок запуска

- `TELEGRAM_BOT_TOKEN is required` — нет переменной в `.env`/окружении.
- Контент не читается — проверьте `CONTENT_PATH` и что вы запускаете процесс из ожидаемой директории.

## Стоп‑правила по разработке

- Если одна и та же ошибка при правках случилась дважды подряд — остановиться и переосмыслить подход.
- Если изменения замкнулись в цикл «та же правка без прогресса» — остановиться и изменить тактику / задать помощь человеку.
