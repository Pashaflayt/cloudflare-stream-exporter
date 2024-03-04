# Cloudflare-stream-exporter

## Собрать докерфайл с минимальным размером с нуля

+ Установить github.com/zeromicro/go-zero/tools/goctl

  ```bash
  brew install goctl
  goctl docker --tz Europe/Moscow --exe cloudflare-stream-exporter -go main.go 
  ```

+ Руками добавить пакер upx в dockerfile

  ```bash
  RUN apk add upx
  RUN upx /cloudflare-stream-exporter 
  ```

## Комментарии по коду main.go

Для сбора и фильтрации метрик используется кастомный коллектор, далее согласно документации:

1. When implementing the collector for your exporter, you should never use the usual direct instrumentation approach and then update the metrics on each scrape. Это значит, что мы не используем циклы и хранение метрики (прямое инструментирование). За редкими исключениями вне этого приложения.
2. If you already have metrics available, created outside of the Prometheus context, you don't need the interface of the various Metric types. You essentially want to mirror the existing numbers into Prometheus Metrics during collection. An own implementation of the Collector interface is perfect for that. You can create Metric instances “on the fly” using NewConstMetric. Метод **MustNewConstMetric** формирует метрики на лету при запросе от сервера Прометеуса.
3. Из экспортера вырезаны стандартные метрики -- согласно документации, не собирать те метрики, которые не будет парсить сервер prometheus. Соответсвенно из конфига джобы убирается drop стандартных метрик.

main.go парсит 2 Gauge значения:\
  **TotalStorageMinutes\
  TotalStorageMinutesLimit**\
Можно парсить и другие значения json, добавив код по аналогии\
Структура данных для всего json уже прописана.

---

В текущем виде docker-compose предназначен для использования с Ansible.

## TODO

1. Реализовать healthcheck, через http сервер отдающий 200 по таймауту. В конфигурации докера добавить проверку хэлзчека/либо через конфиг скрейпа.
2. Интегрировать promlog для запуска в режиме флага debug. В целом это необходимо для более сложных/тяжелых экспортеров, для тюнинга производительсноти
