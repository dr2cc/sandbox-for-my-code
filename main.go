package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type Storage interface {
	Insert(uid string, url string) error
	Get(uid string) (string, error)
	PostHandler(w http.ResponseWriter, req *http.Request)
	GetHandler(w http.ResponseWriter, req *http.Request)
}

// тип urlStorage и его параметр Data
type UrlStorage struct {
	Data map[string]string
}

// конструктор объектов с типом urlStorage
func NewStorageStruct() *UrlStorage {
	return &UrlStorage{
		Data: make(map[string]string),
	}
}

// тип urlStorage и его метод Insert
func (s *UrlStorage) Insert(uid string, url string) error {
	s.Data[uid] = url
	return nil
}

// тип urlStorage и его метод Get
func (s *UrlStorage) Get(uid string) (string, error) {
	e, existss := s.Data[uid]
	if !existss {
		return uid, errors.New("URL with such id doesn`t exist")
	}
	return e, nil
}

// // Создаю запись в передаваемом сюда объекте реализующем интерфейс Storage
// func MakeNewEntry(s Storage, uid string, url string) {
// 	s.Insert(uid, url)
// }

// Функция для генерации сокращённого URL
func generateShortURL(urlList *UrlStorage, longURL string) string {
	// Генерируем уникальный идентификатор (uid) при помощи пакета go.uuid
	uuidObj := uuid.NamespaceURL
	uuidStr := uuidObj.String()
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")
	uid := uuidStr[:8]

	//Вот здесь создаем запись в нашем объекте (типа *UrlStorage)
	//map[string]string ["6ba7b811": "https://practicum.yandex.ru/", ]
	urlList.Insert(uid, longURL)

	return "/" + uid
}

// тип urlStorage и его метод PostHandler
func (ts *UrlStorage) PostHandler(w http.ResponseWriter, req *http.Request) {
	param, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Преобразуем тело запроса (тип []byte) в строку:
	longURL := string(param)
	// Генерируем сокращённый URL и создаем запись в нашем хранилище
	shortURL := req.Host + generateShortURL(ts, longURL)

	// Устанавливаем статус ответа 201
	w.WriteHeader(http.StatusCreated)

	fmt.Fprint(w, shortURL)

}

// тип urlStorage и его метод GetHandler
func (ts *UrlStorage) GetHandler(w http.ResponseWriter, req *http.Request) {
	// //Пока (14.04.2025) не знаю как передать PathValue при тестировании.
	// id := req.PathValue("id")

	// А вот RequestURI получается и от клиента и из теста
	// Но получаем лишний "/"
	id := strings.TrimPrefix(req.RequestURI, "/")

	longURL, err := ts.Get(id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", longURL)
	// //И так и так работает. Оставил первоначальный вариант.
	//http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	//создаем маршрутизатор
	mux := http.NewServeMux()
	//создаю объект типа UrlStorage
	storage := NewStorageStruct()

	//обращаюсь к методам UrlStorage
	mux.HandleFunc("POST /{$}", storage.PostHandler)
	mux.HandleFunc("GET /{id}", storage.GetHandler)
	//mux.Handle("/pizzas", pizzasHandler{&pizzas})

	http.ListenAndServe("localhost:8080", mux)
}
