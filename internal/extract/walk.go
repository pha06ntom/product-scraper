// Walk рекурсивно обходит произвольную JSON-структуру,
// так как структура API может меняться.
//
// Вместо жестко заданных структур,
// мы ищем объекты, содержащие поля,
// похожие на товар (name + price).
package extract

func Walk(v interface{}, onObj func(map[string]interface{})) {
	switch t := v.(type) {
	case map[string]interface{}:
		onObj(t)
		for _, vv := range t {
			Walk(vv, onObj)
		}
	case []interface{}:
		for _, vv := range t {
			Walk(vv, onObj)
		}
	}
}
