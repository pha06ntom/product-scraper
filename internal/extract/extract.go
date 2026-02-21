// Package extract отвечает за извлечение товарных данных
// из произвольных JSON-ответов backend API.
//
// Мы НЕ парсим HTML-разметку страницы,
// а анализируем JSON (network responses), что делает решение
// более устойчивым к изменениям верстки.
package extract

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Item — минимальный набор данных о товаре,
// который мы хотим выгрузить.
type Item struct {
	Name  string
	Price string
	URL   string
}

// регулярка для строк, похожих на цену
var rePriceLike = regexp.MustCompile(`^\d+([.,]\d+)?$`)

// FromAnyJSON пытается найти в произвольной JSON-структуре
// объекты, похожие на "товар": есть name/title + price.
// respURL — это URL запроса, из которого пришёл этот JSON;
// используется для восстановления базового URL товара.
func FromAnyJSON(v interface{}, respURL string) []Item {
	var out []Item

	// парсим базовый URL (scheme + host), чтобы потом собрать ссылку
	u, _ := url.Parse(respURL)
	base := ""
	if u != nil {
		base = u.Scheme + "://" + u.Host
	}

	Walk(v, func(obj map[string]interface{}) {
		// Название товара: пробуем несколько ключей
		name := pickString(obj, "name", "title", "productName", "displayName")
		if name == "" {
			return
		}

		// Цена товара
		price := pickPrice(obj)
		if price == "" {
			return
		}

		// Ссылка на товар
		link := pickString(obj, "url", "link", "productUrl", "href", "slug")

		// Если ссылка относительная, достраиваем от базового хоста
		if link != "" && strings.HasPrefix(link, "/") && base != "" {
			link = base + link
		}

		// Если явной ссылки нет, пробуем собрать её из id/sku
		if link == "" {
			if id := pickString(obj, "id", "productId", "code", "sku"); id != "" && base != "" {
				link = base + "/search/?q=" + url.QueryEscape(id)
			}
		}

		// В крайнем случае — ссылка на сам API-URL,
		// чтобы хотя бы была точка входа к информации.
		if link == "" {
			link = respURL
		}

		out = append(out, Item{
			Name:  normalizeSpace(name),
			Price: price,
			URL:   link,
		})
	})

	return out
}

// pickString возвращает первое непустое строковое поле
// из перечисленных ключей.
func pickString(obj map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if val, ok := obj[k]; ok {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return ""
}

// pickPrice пытается вытащить цену из объекта.
// Сначала смотрим по набору "типичных" ключей,
// потом по вложенным объектам, потом по всем полям, содержащим "price".
func pickPrice(obj map[string]interface{}) string {
	// 1. Фиксированные ключи на верхнем уровне
	for _, k := range []string{"price", "currentPrice", "regularPrice", "value", "amount", "salePrice", "priceValue"} {
		if val, ok := obj[k]; ok {
			if s := toPriceString(val); s != "" {
				return s
			}
		}
	}

	// 2. Часто цена лежит во вложенном объекте price/prices
	for _, k := range []string{"price", "prices"} {
		if val, ok := obj[k]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				// сначала все поля, содержащие "price"
				for key, vv := range m {
					if strings.Contains(strings.ToLower(key), "price") {
						if s := toPriceString(vv); s != "" {
							return s
						}
					}
				}
				// затем типичные варианты
				for _, kk := range []string{"value", "current", "regular", "amount"} {
					if s := toPriceString(m[kk]); s != "" {
						return s
					}
				}
			}
		}
	}

	// 3. Fallback: любой ключ с подстрокой "price"
	for key, val := range obj {
		if strings.Contains(strings.ToLower(key), "price") {
			if s := toPriceString(val); s != "" {
				return s
			}
		}
	}

	return ""
}

// toPriceString приводит значение к строке-цене, если это возможно.
func toPriceString(v interface{}) string {
	switch t := v.(type) {
	case float64:
		if t <= 0 {
			return ""
		}
		// если целое — без дробной части
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return strconv.FormatFloat(t, 'f', -1, 64)

	case string:
		s := strings.TrimSpace(strings.ReplaceAll(t, "₽", ""))
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\u00a0", "")
		if s == "" {
			return ""
		}
		s = strings.ReplaceAll(s, ",", ".")
		if rePriceLike.MatchString(s) {
			return s
		}

	case map[string]interface{}:
		// иногда цена лежит в подполе value/amount
		if vv, ok := t["value"]; ok {
			return toPriceString(vv)
		}
		if vv, ok := t["amount"]; ok {
			return toPriceString(vv)
		}
	}

	return ""
}

// normalizeSpace удаляет лишние пробелы и переводит
// все последовательности whitespace в одиночный пробел.
func normalizeSpace(s string) string {
	s = strings.TrimSpace(s)
	return strings.Join(strings.Fields(s), " ")
}
