package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/pha06ntom/lenta-scraper/internal/extract"
)

type Collector struct {
	mu   sync.Mutex
	seen map[string]struct{}
	out  []extract.Item
}

func NewCollector() *Collector {
	return &Collector{
		seen: make(map[string]struct{}),
	}
}

// Attach подключает перехват JSON-ответов backend API через CDP Network events.
func (c *Collector) Attach(ctx context.Context) {
	// Включаем сетевой домен CDP.
	_ = chromedp.Run(ctx, network.Enable())

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		e, ok := ev.(*network.EventResponseReceived)
		if !ok {
			return
		}

		// Берём только ответы с JSON в content-type.
		ct := headerGet(e.Response.Headers, "content-type")
		if !strings.Contains(strings.ToLower(ct), "json") {
			return
		}

		respURL := e.Response.URL

		u := strings.ToLower(respURL)
		if !(strings.Contains(u, "kuper.ru") || strings.Contains(u, "api")) {
		}
		// Забираем тело ответа асинхронно.
		go func(reqID network.RequestID, respURL string) {
			body, err := network.GetResponseBody(reqID).Do(ctx)
			if err != nil || len(body) == 0 {
				return
			}

			var anyJSON interface{}
			if err := json.Unmarshal(body, &anyJSON); err != nil {
				return
			}

			items := extract.FromAnyJSON(anyJSON, respURL)
			if len(items) == 0 {
				return
			}

			c.mu.Lock()
			defer c.mu.Unlock()

			for _, it := range items {
				if it.Name == "" || it.Price == "" || it.URL == "" {
					continue
				}
				key := fmt.Sprintf("%s|%s|%s", it.Name, it.Price, it.URL)
				if _, ok := c.seen[key]; ok {
					continue
				}
				c.seen[key] = struct{}{}
				c.out = append(c.out, it)
			}
		}(e.RequestID, respURL)
	})
}

func (c *Collector) Items() []extract.Item {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]extract.Item, len(c.out))
	copy(cp, c.out)
	return cp
}

func headerGet(h network.Headers, key string) string {
	key = strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == key {
			return fmt.Sprint(v)
		}
	}
	return ""
}
