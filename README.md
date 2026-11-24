# Мини-спортсбук

Запуск: `docker compose up --build`, затем http://localhost:8080. Начальный баланс 1000 ₽. Postgres хранит данные в volume; `docker compose down -v` удалит их.

## Линия

Два независимых случайных источника генерируют вероятности для шести матчей. Для каждого источника вероятности нормализуются, затем усредняются. Наш коэффициент = floor(100 / (p \* 1.05)) / 100. Overround минимум 5%, округление вниз немного повышает его. Обновление раз в 5 секунд, доставка через SSE с автоматическим переподключением. Линия хранится в памяти: рассчитано на один экземпляр API.

## Ставки и деньги

Все суммы — целые копейки bigint, коэффициенты — целые сотые. Клиент передаёт сумму копеек строкой. Транзакция блокирует единственную строку кошелька FOR UPDATE, проверяет повтор ключа, коэффициент и достаточность средств; запись ставки и списание атомарны. CHECK защищает от отрицательного баланса. Повтор с тем же ключом и параметрами не списывает деньги повторно. При неопределённом результате интерфейс сохраняет запрос для повтора. Изменённый коэффициент отклоняется, автоматического согласия нет.

## Проверки и ограничения

`cd backend && go test ./...`; `cd frontend && npm install && npm run build`. Тест проверяет границы маржи на случайных линиях. TypeScript strict и vue-tsc включены. Docker compose config проверен. Полный запуск контейнеров в среде агента недоступен: доступ к Docker pipe запрещён. Проверки сборок зависят от доступности загрузки пакетов.

Следующие шаги: интеграционные тесты Postgres для конкурентных списаний, повторов и rollback; история ставок; восстановление неопределённого запроса после перезагрузки через localStorage; серверные таймауты и мониторинг. Нет авторизации (один демонстрационный игрок), расчёта ставок и админки согласно заданию. Не публиковать демонстрацию как многопользовательский сервис.

## Авторство и время

ИИ создал весь код, конфигурацию, тест и README в этой сессии. Пользователь предоставил задание; ручную работу пользователя над кодом агент не наблюдал. Работа выполнена за одну короткую сессию, менее 3 часов; точное время следует указать по собственному учёту перед сдачей. Репозиторий и публикация остаются на стороне владельца.

## Linting

Install tooling: `npm install` in the repository root and in `frontend`. Run `pwsh -File scripts/lint.ps1` after every change. Use `-Fix` to format. Checks cover ESLint (TypeScript/Vue), Stylelint (CSS), Prettier (HTML/JSON/YAML/Markdown), Go formatting/vet/tests, TypeScript build, Compose validation, Hadolint (Dockerfiles), and nginx configuration validation. Docker must be running; nginx validation uses the nginx:alpine image. All indentation is four spaces; Go uses gofmt output with tabs expanded to four spaces.

PowerShell lint setup (once): `Save-Module -Name PSScriptAnalyzer -RequiredVersion 1.24.0 -Path .cache/psmodules -Force`.
