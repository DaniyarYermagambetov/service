package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type MemStorage struct {
	Counters map[string]int64
	Gauges   map[string]float64
	mu       sync.Mutex
}

func initMemStorage() *MemStorage {
	return &MemStorage{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}
}

// Метод для обновления counter
func (s *MemStorage) IncrementCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Counters[name] += value
}

// Метод для обновления gauge
func (s *MemStorage) SetGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Gauges[name] = value
}

var storage = initMemStorage()

func main() {
	// Создаем маршрутизатор
	mux := http.NewServeMux()

	// Регистрируем обработчик для обновления метрик
	mux.HandleFunc("/update/", updateHandler)

	// Запускаем сервер
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

// Обработчик для обновления метрик
func updateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод запроса POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Разбиваем путь URL
	urlPath := r.URL.Path
	splitedUrlPath := strings.Split(urlPath, "/")

	// Если путь URL не соответствует ожидаемому формату
	if len(splitedUrlPath) != 5 {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	// Извлекаем параметры из пути
	typeMetric := splitedUrlPath[2]
	nameMetric := splitedUrlPath[3]
	valueMetric := splitedUrlPath[4]

	// Обрабатываем тип метрики
	switch typeMetric {
	case "counter":
		value, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, "Incorrect metric value! Must be int64", http.StatusBadRequest)
			return
		}
		storage.IncrementCounter(nameMetric, value)

	case "gauge":
		value, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, "Incorrect metric value! Must be float64", http.StatusBadRequest)
			return
		}
		storage.SetGauge(nameMetric, value)

	default:
		http.Error(w, "Undefined metric type!", http.StatusBadRequest)
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
