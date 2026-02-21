package output

import (
	"encoding/csv"
	"os"

	"github.com/pha06ntom/lenta-scraper/internal/extract"
)

func WriteCSV(path string, items []extract.Item) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"name", "price", "url"}); err != nil {
		return err
	}

	for _, it := range items {
		if err := w.Write([]string{it.Name, it.Price, it.URL}); err != nil {
			return err
		}
	}
	return w.Error()
}
