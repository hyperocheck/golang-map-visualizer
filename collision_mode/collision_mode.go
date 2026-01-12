package collision_mode

import (
	"fmt"
	"unsafe"

	"visualizer/src/engine"
	"visualizer/src/logger"
)

var (
    mapType *_maptype_
)

func CheckHash[K comparable, V any](t *engine.Type[K, V], key K) uint8 {
	
	if mapType == nil {
		addr := GetMapType(t.Data)
		mapType = (*_maptype_)(unsafe.Pointer(addr))
	}

	hmap := t.GetHmap()

	var hf hashFunc
	*(*uintptr)(unsafe.Pointer(&hf)) = *(*uintptr)(unsafe.Pointer(&mapType.Hasher))
	
	keyPtr := unsafe.Pointer(&key)
	seed := uintptr(hmap.Hash0) 
	hash := hf(keyPtr, seed)

	//top := tophash(hash) 
	logger.Log.Log("INFO", fmt.Sprintf("key: %d, tophash: %d", key, lowerbytes(hash, hmap.B)))

	return lowerbytes(hash, hmap.B)
}

func lowerbytes(hash uintptr, B uint8) uint8 {
	return uint8(hash & ((uintptr(1) << B) - 1))
}

func tophash(hash uintptr) uint8 {
	top := uint8(hash >> (8 * (unsafe.Sizeof(uintptr(0)) - 1)))
	if top < 5 { 
		top += 5
	}
	return top
}

/*
func main() {
	// 1. Инициализируем карту и данные
	key := 9999999999
	val := 4444444444
	m := map[int]int{key: val}

	// 2. Получаем метаданные типа через наш ассемблерный хук
	// Передаем саму карту в интерфейс
	addr := GetMapType(m)
	mt := (*_maptype_)(unsafe.Pointer(addr))

	// 3. Получаем доступ к заголовку hmap
	// m в Go — это указатель на hmap, поэтому берем адрес переменной m
	hmap := *(**Hmap)(unsafe.Pointer(&m))

	// 4. Подготавливаем функцию хеширования
	var hf hashFunc
	*(*uintptr)(unsafe.Pointer(&hf)) = *(*uintptr)(unsafe.Pointer(&mt.Hasher))

	// 5. Считаем хеш так же, как это делает рантайм
	keyPtr := unsafe.Pointer(&key)
	seed := uintptr(hmap.hash0) // Берем seed именно из этой карты
	hash := hf(keyPtr, seed)
	calcTop := tophash(hash)

	// 6. Читаем данные из первого бакета
	bPointer := (*_bucket_[int, int])(hmap.buckets)

	// --- Красивый вывод ---
	fmt.Printf("--- Отладка Map (Go 1.23.12) ---\n")
	fmt.Printf("Seed карты (hash0):  0x%x\n", hmap.hash0)
	fmt.Printf("Полный хеш ключа:   0x%016x\n", hash)
	fmt.Printf("Наш расчетный Top:  0x%02x (%d)\n", calcTop, calcTop)
	
	fmt.Print("Бакет Tophash:      [ ")
	for _, b := range bPointer.tophash {
		fmt.Printf("0x%02x ", b)
	}
	fmt.Println("]")

	// Проверка
	match := false
	for i := 0; i < 8; i++ {
		if bPointer.tophash[i] == calcTop {
			fmt.Printf("РЕЗУЛЬТАТ: Найдено совпадение в слоте %d!\n", i)
			fmt.Printf("Значение в слоте:   %d\n", bPointer.values[i])
			match = true
			break
		}
	}

	if !match {
		fmt.Println("РЕЗУЛЬТАТ: Совпадение не найдено. Проверьте структуры.")
	}
}
*/

