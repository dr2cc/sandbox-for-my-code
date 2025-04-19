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
	GetURL(uid string) (string, error)
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
func (s *UrlStorage) GetURL(uid string) (string, error) {
	e, existss := s.Data[uid]
	if !existss {
		return uid, errors.New("URL with such id doesn`t exist")
	}
	return e, nil
}

// Реализую интерфейс Storage - создаю запись в передаваемом сюда объекте
func MakeNewEntry(s Storage, uid string, url string) {
	s.Insert(uid, url)
}

// Функция для генерации сокращённого URL
func generateShortURL(urlList *UrlStorage, longURL string) string {
	// Генерируем уникальный идентификатор (uid) при помощи пакета go.uuid
	uuidObj := uuid.NamespaceURL
	uuidStr := uuidObj.String()
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")
	uid := uuidStr[:8]

	// //Вот здесь создаем запись в нашем объекте (типа *UrlStorage)
	// //map[string]string ["6ba7b811": "https://practicum.yandex.ru/", ]
	// //Так можно вызвать метод Insert не реализуя интерфейс Storage
	//urlList.Insert(uid, longURL)

	//Реализуем интерфейс Storage, что в последующем даст возможность
	//использовать его методы и другим типам
	MakeNewEntry(urlList, uid, longURL)

	return "/" + uid
}

// тип urlStorage и его метод PostHandler
func (ts *UrlStorage) PostHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		switch req.Header.Get("Content-Type") {
		case "text/plain":
			param, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Преобразуем тело запроса (тип []byte) в строку:
			longURL := string(param)
			// Генерируем сокращённый URL и создаем запись в нашем хранилище
			shortURL := req.Host + generateShortURL(ts, longURL)

			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, shortURL)
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Content-Type isn`t text/plain")
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Method not allowed")
	}
}

// Реализую интерфейс Storage - получаю запись из объекта хранилища
func GetEntry(s Storage, uid string) (string, error) {
	e, err := s.GetURL(uid)
	return e, err
}

// тип urlStorage и его метод GetHandler
func (ts *UrlStorage) GetHandler(w http.ResponseWriter, req *http.Request) {
	//Тесты подсказали добавить проверку на метод:
	switch req.Method {
	case http.MethodGet:
		// //Пока (14.04.2025) не знаю как передать PathValue при тестировании.
		// id := req.PathValue("id")

		// А вот RequestURI получается и от клиента и из теста
		// Но получаем лишний "/"
		id := strings.TrimPrefix(req.RequestURI, "/")

		// //Так не реализуя интерфейс
		//longURL, err := ts.GetURL(id)

		//Так реализуя интерфейс
		longURL, err := GetEntry(ts, id)
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
		w.WriteHeader(http.StatusBadRequest)
	}
}

// *********************************************************************************
// CustomMux переопределяет стандартный ServeMux, чтобы вместо 405 возвращать 400
type CustomMux struct {
	*http.ServeMux
}

func (m *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Проверяем, есть ли такой путь
	_, pattern := m.Handler(r)
	if pattern == "" {
		// // Если эндпоинта нет вообще — 404
		// http.NotFound(w, r)

		// Но мне нужно 400
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Если эндпоинт есть, но метод не совпадает — 400
	if !m.isMethodAllowed(r) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Иначе передаем обработку стандартному ServeMux
	m.ServeMux.ServeHTTP(w, r)
}

// isMethodAllowed проверяет, разрешен ли метод для данного пути
func (m *CustomMux) isMethodAllowed(r *http.Request) bool {
	// Получаем зарегистрированный обработчик для этого пути
	handler, _ := m.Handler(r)

	// Если обработчик — это ServeMux (значит, метод не совпадает)
	_, isServeMux := handler.(*http.ServeMux)
	return !isServeMux
}

func main() {
	// mux := http.NewServeMux()

	//Для создания ответ 400 на все не верные запросы
	//создаю кастомный ServeMux (маршрутизатор)
	mux := &CustomMux{http.NewServeMux()}

	//создаю объект типа UrlStorage
	storage := NewStorageStruct()

	//обращаюсь к методам UrlStorage
	mux.HandleFunc("POST /{$}", storage.PostHandler)
	mux.HandleFunc("GET /{id}", storage.GetHandler)

	http.ListenAndServe("localhost:8080", mux)
}
