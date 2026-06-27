# go-solutions

Коллекция переиспользуемых Go-пакетов для высоконагруженных систем, математики, машинного обучения, безопасности и многого другого.

## 🇬🇧 Documentation / 🇷🇺 Документация

- [📖 Package Reference / Справочник пакетов](docs/index.md) – полный список всех модулей и их методов с примерами и диаграммами.

## Быстрый поиск

Используйте `Ctrl+F` по ключевым словам:  
`деньги`, `LTV`, `PCA`, `сортировка`, `растояние`, `JWT`, `токен`, `кэш`, `rate limit`, `User-Agent`, `геолокация`, `орбита`, `Кеплер`, `N-body`, `тензор`, `нейросеть`, `Adam`, `ReLU` …

Или смотрите [Алфавитный указатель](#алфавитный-указатель) внизу.

## Модули

### math
- **`math/fixedpoint`** – точная денежная арифметика (копейки/центы)
- **`math/statistics`** – среднее, дисперсия, корреляция, перцентили
- **`math/regression`** – линейная и логистическая регрессия
- **`math/factoranalysis`** – метод главных компонент (PCA)
- **`math/timeseries`** – прогнозирование временных рядов (Хольт-Уинтерс)
- **`math/radixsort`** – поразрядная сортировка (LSD) за O(N)
- **`math/geometry`** – векторы, матрицы, кватернионы (планируется)
- **`math/optimization`** – градиентный спуск, симплекс-метод (планируется)
- **`math/random`** – ГПСЧ, шум Перлина, сэмплирование (планируется)
- **`math/ml`** – нейронные сети, оптимизаторы, функции потерь (планируется)
- **`math/highprec`** – высокоточная арифметика (big.Int/Float/Rat) и научные расчёты (планируется)

### net
- **`net/backpressure`** – конвейер с обратным давлением
- **`net/breaker`** – circuit breaker
- **`net/ratelimit`** – ограничитель частоты (token bucket)

### data
- **`data/timedcache`** – кэш с TTL и финализатором
- **`data/idempotent`** – хранилище ключей идемпотентности
- **`data/sampler`** – вероятностный сэмплер
- **`data/experiment`** – A/B-флаги
- **`data/registry`** – типобезопасный реестр обработчиков

### security
- **`security/appsec`** – безопасные редиректы, санитизация HTML, HMAC
- **`security/geoip`** – GeoIP (MaxMind)
- **`security/device`** – парсинг User-Agent

### geo
- **`geo/geospatial`** – расстояние Хаверсина, точка в полигоне

### valuation
- **`valuation`** – LTV, ценность показа, гео-фактор, оптимальная ставка

### config
- **`config`** – загрузка YAML и переменных окружения

### logger
- **`logger`** – структурированный логгер на базе `slog`

### metrics
- **`metrics`** – метрики OpenTelemetry

### shutdown
- **`shutdown`** – корректное завершение с приоритетами

---

## Алфавитный указатель

| Ключевые слова | Пакет |
|----------------|-------|
| A/B, эксперимент | `data/experiment` |
| big.Float, высокая точность, астрофизика | `math/highprec` |
| circuit breaker | `net/breaker` |
| distance, Haversine, polygon | `geo/geospatial` |
| JWT, токен | (пока нет) |
| LTV, ценность показа, гео-фактор | `valuation` |
| PCA, факторный анализ | `math/factoranalysis` |
| radix sort, сортировка LSD | `math/radixsort` |
| rate limit, ограничение частоты | `net/ratelimit` |
| regression, регрессия | `math/regression` |
| timed cache, кэш с TTL | `data/timedcache` |
| user-agent, парсинг | `security/device` |
| деньги, копейки | `math/fixedpoint` |
| ... | ... |
