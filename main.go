package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Product struct {
	Name  string `json:"name"`
	Price string `json:"price"`
	URL   string `json:"url"`
}

func main() {
	/*opts := append(chromedp.DefaultExecAllocatorOptions[:],
	    chromedp.Flag("headless", true),
	    chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)*/

	allocCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222/")
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	MoscowCookie := map[string]string{
		"App_Cache_CitySlug": "moscow",
		"App_Cache_City":     `%7B%22centerLat%22%3A%2255.75322000%22%2C%22centerLng%22%3A%2237.62255200%22%2C%22id%22%3A1%2C%22isDefault%22%3Atrue%2C%22mainDomain%22%3Afalse%2C%22name%22%3A%22%D0%9C%D0%BE%D1%81%D0%BA%D0%B2%D0%B0%20%D0%B8%20%D0%9C%D0%9E%22%2C%22slug%22%3A%22moscow%22%7D`,
	}

	var products []Product
	categories := []string{
		"https://lenta.com/catalog/molochnye-produkty-3/",
		"https://lenta.com/catalog/syry-2/",
	}

	for _, catURL := range categories {
		fmt.Printf("🔍 Парсинг: %s\n", catURL)

		var names, prices, links []string

		err := chromedp.Run(ctx,
			chromedp.EmulateViewport(1920, 1080),
			chromedp.ActionFunc(func(ctx context.Context) error {
				for k, v := range MoscowCookie {
					if err := network.SetCookie(k, v).WithDomain(".lenta.com").WithPath("/").Do(ctx); err != nil {
						return err
					}
				}
				return nil
			}),

			chromedp.Navigate(catURL),

			//chromedp.WaitVisible(`.sku-card-small-container, .product-card`, chromedp.ByQuery),
			//chromedp.WaitVisible(`.sku-card-small-container, .product-card-container, [data-testid="sku-card"]`, chromedp.ByQuery),
			chromedp.Sleep(5*time.Second),
			// Цикла для того чтобы показать товары не только на первой странице
			chromedp.ActionFunc(func(ctx context.Context) error {
				maxClicks := 50

				for i := 0; i < maxClicks; i++ {
					fmt.Printf("Страница %d: Скроллим вниз...\n", i+1)

					if err := chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil).Do(ctx); err != nil {
						return err
					}

					time.Sleep(2 * time.Second)

					var found bool
					err := chromedp.Evaluate(`(function(){
                        let btns = Array.from(document.querySelectorAll('button'));
                        let btn = btns.find(b => b.innerText.includes('Показать ещё'));
                        
                        if (btn) {
                            btn.scrollIntoView({block: "center", behavior: "auto"});
                            btn.click();
                            return true;
                        }
                        return false;
                    })()`, &found).Do(ctx)

					if err != nil {
						fmt.Printf("❌ Ошибка JS: %v\n", err)
						break
					}

					if !found {
						fmt.Println("Кнопка 'Показать ещё' не появилась.")
						break
					}

					time.Sleep(3 * time.Second)
				}
				return nil
			}),

			chromedp.Evaluate(`
                Array.from(document.querySelectorAll('a[href*="/product/"]')).map(e => e.innerText.trim())
            `, &names),

			chromedp.Evaluate(`
                Array.from(document.querySelectorAll('*')).filter(e => e.innerText && e.innerText.includes('₽') && e.children.length === 0).map(e => e.innerText.trim())
            `, &prices),

			chromedp.Evaluate(`
                Array.from(document.querySelectorAll('a[href*="/product/"]')).map(e => e.href)
            `, &links),
		)

		if err != nil {
			log.Printf("❌ Ошибка %s: %v", catURL, err)
		}

		limit := len(names)
		if len(prices) < limit {
			limit = len(prices)
		}

		for i := 0; i < limit; i++ {
			rawLines := strings.Split(names[i], "\n")
			cleanName := ""

			for _, line := range rawLines {
				line = strings.TrimSpace(line)

				// фильтрация для названия продукта
				if len(line) > 3 &&
					!strings.ContainsAny(line[:1], "0123456789") &&
					!strings.Contains(line, "Баллы за отзыв") &&
					!strings.Contains(line, "Цена за 1 шт") &&
					!strings.Contains(line, "С Картой №1") &&
					!strings.Contains(line, "WOW-находка") &&
					!strings.Contains(line, "Сделано в Беларус") &&
					!strings.Contains(line, "Наша марка") &&
					!strings.Contains(line, "Местный продукт") &&
					!strings.Contains(line, "Выгодно") {

					cleanName = line
					break
				}
			}

			if cleanName == "" && len(rawLines) > 0 {
				cleanName = rawLines[0]
			}
			p := Product{
				Name:  cleanName,
				Price: strings.TrimSpace(prices[i]),
				URL:   links[i],
			}
			products = append(products, p)
			fmt.Printf("✅ %s - %s\n", p.Name+"...", p.Price)
		}
	}

	if len(products) > 0 {
		data, _ := json.MarshalIndent(products, "", " ")
		os.WriteFile("products.json", data, 0644)
		fmt.Printf("\n🎉 Успешно! %d товаров сохранено.\n", len(products))
	} else {
		fmt.Println("\n!!! Товары не найдены!!!")
	}
}
