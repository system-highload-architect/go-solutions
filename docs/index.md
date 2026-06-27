# 📖 Package Reference / Справочник пакетов

Этот справочник содержит подробное описание каждого пакета библиотеки **go‑solutions**: назначение, методы, меры предосторожности и диаграммы.

## 🧭 Навигация

| Раздел | Пакеты |
|--------|--------|
| **Базовые** | [config](config.md), [logger](logger.md), [metrics](metrics.md), [shutdown](shutdown.md), [zerocopy](zerocopy.md) |
| **Математика** | [fixedpoint](math/fixedpoint.md), [statistics](math/statistics.md), [regression](math/regression.md), [factoranalysis](math/factoranalysis.md), [timeseries](math/timeseries.md), [radixsort](math/radixsort.md) |
| **Данные** | [timedcache](data/timedcache.md), [lru](data/lru.md), [timewheel](data/timewheel.md), [idempotent](data/idempotent.md), [sampler](data/sampler.md), [experiment](data/experiment.md), [registry](data/registry.md) |
| **Сеть** | [backpressure](net/backpressure.md), [breaker](net/breaker.md), [ratelimit](net/ratelimit.md) |
| **Безопасность** | [appsec](security/appsec.md), [geoip](security/geoip.md), [device](security/device.md) |
| **Гео** | [geospatial](geo/geospatial.md) |
| **Оценка ставок** | [valuation](valuation.md) |

## 📝 Как читать описания

Каждый пакет документирован по единому шаблону:
- **Назначение** — для каких задач создан пакет.
- **Основные типы и методы** — ключевые экспортируемые сущности.
- **Меры предосторожности** — на что обратить внимание (например, unsafe‑преобразования, конкурентный доступ).
- **Пример использования** — ссылка на `example/main.go` внутри пакета.
- **Диаграмма** (если применимо) — mermaid‑схема алгоритма или потока данных.