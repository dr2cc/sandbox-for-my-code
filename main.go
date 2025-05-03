package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Storager interface {
	InsertURL(uid string, url string) error
	GetURL(uid string) (string, error)
}

// тип URLStorage и его параметр Data
type URLStorage struct {
	Data map[string]string
}

// конструктор объектов с типом URLStorage
func NewStorageStruct() *URLStorage {
	return &URLStorage{
		Data: make(map[string]string),
	}
}

// тип *URLStorage и его метод InsertURL
func (s *URLStorage) InsertURL(uid string, url string) error {
	s.Data[uid] = url
	return nil
}

// тип *URLStorage и его метод GetURL
func (s *URLStorage) GetURL(uid string) (string, error) {
	e, existss := s.Data[uid]
	if !existss {
		return uid, errors.New("URL with such id doesn`t exist")
	}
	return e, nil
}

//*******************************************************************
// Реализую интерфейс Storage

func MakeEntry(s Storager, uid string, url string) {
	s.InsertURL(uid, url)
}

func GetEntry(s Storager, uid string) (string, error) {
	e, err := s.GetURL(uid)
	return e, err
}

//********************************************************************

func generateShortURL(urlList *URLStorage, longURL string) string {
	// Инициализация генератора случайных чисел
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	runes := []rune(longURL)
	r.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})
	//удаляю из полученной строки все кроме букв и цифр
	reg := regexp.MustCompile(`[^a-zA-Zа-яА-Я0-9]`)
	//[:11] здесь сокращаю строку
	id := reg.ReplaceAllString(string(runes[:11]), "")

	//Реализую интерфейс Storage, что в последующем даст возможность
	//использовать его методы и другим типам
	MakeEntry(urlList, id, longURL)

	return "/" + id
}

// тип *URLStorage и его метод PostHandler
func (ts *URLStorage) PostHandler(w http.ResponseWriter, req *http.Request) {
	// Автотесты не проходили на еще одном уровне switch
	//не знаю как на этом, но без него проходит любой тип контента,
	//а возвратиться может только text
	switch req.Header.Get("Content-Type") {
	case "text/plain":
		//param - тело запроса (тип []byte)
		param, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Генерирую ответ и создаю запись в хранилище
		response := "http://" + req.Host + generateShortURL(ts, string(param))

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, response)
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Content-Type isn`t text/plain")
	}
}

// тип *URLStorage и его метод GetHandler
func (ts *URLStorage) GetHandler(w http.ResponseWriter, req *http.Request) {
	//Тесты подсказали добавить проверку на метод:
	switch req.Method {
	case http.MethodGet:
		// //Пока (14.04.2025) не знаю как передать PathValue при тестировании.
		// id := req.PathValue("id")

		// А вот RequestURI получается и от клиента и из теста
		// Но получаю лишний "/"
		id := strings.TrimPrefix(req.RequestURI, "/")

		//Реализую интерфейс
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

func main() {
	mux := http.NewServeMux()

	//создаю объект типа *URLStorage
	storage := NewStorageStruct()

	//обращаюсь к методам *URLStorage
	mux.HandleFunc("POST /{$}", storage.PostHandler)
	mux.HandleFunc("GET /{id}", storage.GetHandler)

	http.ListenAndServe("localhost:8081", mux)
}
