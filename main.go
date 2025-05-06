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
func (s *UrlStorage) PostHandler(w http.ResponseWriter, req *http.Request) {
	param, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Преобразуем тело запроса (тип []byte) в строку:
	longURL := string(param)
	// Генерируем сокращённый URL и создаем запись в нашем хранилище
	shortURL := "http://" + req.Host + generateShortURL(s, longURL)

	// Устанавливаем статус ответа 201
	w.WriteHeader(http.StatusCreated)

	fmt.Fprint(w, shortURL)

}

// тип urlStorage и его метод GetHandler
func (s *UrlStorage) GetHandler(w http.ResponseWriter, req *http.Request) {
	//Тесты подсказали добавить проверку на метод:
	switch req.Method {
	case http.MethodGet:
		// //Пока (14.04.2025) не знаю как передать PathValue при тестировании.
		// id := req.PathValue("id")

		// А вот RequestURI получается и от клиента и из теста
		// Но получаем лишний "/"
		id := strings.TrimPrefix(req.RequestURI, "/")

		longURL, err := s.Get(id)
		if err != nil {
			//http.Error(w, "URL not found", http.StatusBadRequest)
			w.Header().Set("Location", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", longURL)
		// //И так и так работает. Оставил первоначальный вариант.
		//http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.Header().Set("Location", "Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	// Любая функция удовлетворяющая интерфейсу Handler будет реализовывать его.
	//
	//	Handler interface {
	//	   ServeHTTP(ResponseWriter, *Request)}
	//
	// Как я понимаю выбор для встроенных данных типа interface имеет смыслом возможность
	// создания бесконечности собственных функций удовлетворяющих ему.
	//
	// You could not pass it as a handler just yet. After all, it’s just a function and does not provide that ServeHTTP method signature.
	// You can use HandlerFunc to wrap this function, and have it satisfy the interface!
	// https://gopherdojo.com/handler-vs-handlerfunc/
	//

	// создаю маршрутизатор
	mux := http.NewServeMux()
	// создаю объект типа UrlStorage
	storage := NewStorageStruct()
	//

	// Не путать две сигнатуры в пакете http - HandleFunc и HandlerFunc (!!!)
	//
	// HandleFunc является методом типа ServeMux
	// func (mux *http.ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	// or
	// http.HandleFunc(pattern string, handler func(ResponseWriter, *Request))
	//
	// Метод HandleFunc регистрирует функцию-обработчик для заданного шаблона.
	// Если данный шаблон конфликтует с уже зарегистрированным, HandleFunc в панике.
	//
	// HandlerFunc является типом
	// type HandlerFunc func(ResponseWriter, *Request)
	// Тип HandlerFunc — это адаптер (обертка), позволяющий использовать обычные функции (но реализующие Handler interface)
	// в качестве обработчиков [Handler] HTTP.
	// Если f — это функция с соответствующей сигнатурой, HandlerFunc(f) — это [Handler], который вызывает f.
	// The HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
	// If f is a function with the appropriate signature, HandlerFunc(f) is a [Handler] that calls f.
	//
	// А вот ServeHTTP это метод типа HandlerFunc
	// он и вызывает нашу обычную функцию (в примере выше f)
	// ServeHTTP calls f(w, r).
	//
	// func (f http.HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request)

	// обращаюсь к методам UrlStorage
	// но,
	// через метод HandlerFunc для обертывания удовлетворяющей интерфейсу Handler
	// func (ts *UrlStorage) PostHandler(w http.ResponseWriter, req *http.Request) {}
	// (она реализует его методы- ResponseWriter и Request)
	// Без этого ее нельзя передать как обработчик т.к. она не предоставляет сигнатуру метода ServeHTTP.
	mux.HandleFunc("POST /{$}", storage.PostHandler)
	mux.HandleFunc("GET /{id}", storage.GetHandler)
	//
	http.HandleFunc("POST /{$}", storage.PostHandler)
	//func http.HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	// http.HandlerFunc
	// //type HandlerFunc func(ResponseWriter, *Request)

	http.ListenAndServe("localhost:8080", mux)
}
