# Lenta Catalog Parser (chromedp + Remote Chrome)

Мини программка для парсинга каталога товаров сети **«Лента»** через уже запущенный браузер Chrome по протоколу DevTools.  
Посде обработки сохраняются результаты в `products.json`. 

---

## Краткая информация

- Подключение к существующему экземпляру Chrome по WebSocket (`ws://127.0.0.1:9222/`). 
- Парсинг нескольких категорий каталога Ленты (молочная продукция и сыры). 
- Автоклик по кнопке «Показать ещё» с прокруткой страницы для загрузки всех товаров категории.
- Получение для каждого товара:
  - названия,
  - цены (с `₽`),
  - ссылки на карточку товара.  
  - Выгрузка данных в:
  - `products.json`.

---

### Модуль и зависимости

```bash
go mod init test-lenta
go get github.com/chromedp/chromedp
go get github.com/chromedp/cdproto/network
```

### Запуск Chrome с remote debugging

Для работы парсира используем отдельное окно Chrome.

#### macOS:
```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --remote-debugging-port=9222 \
  --user-data-dir="/tmp/chrome_remote_profile"
```

#### Linux
```bas
google-chrome \
  --remote-debugging-port=9222 \
  --user-data-dir="/tmp/chrome_remote_profile"
```

#### Windows
```bash
"& 'C:\Program Files\Google\Chrome\Application\chrome.exe'" `
  --remote-debugging-port=9222 `
  --user-data-dir="C:\chrome_remote_profile"
```
