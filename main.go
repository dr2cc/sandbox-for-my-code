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

func (s *UrlStorage) Insert(uid string, url string) error {
	s.Data[uid] = url
	return nil
}

func (s *UrlStorage) Get(uid string) (string, error) {
	e, existss := s.Data[uid]
	if !existss {
		return uid, errors.New("URL with such id doesn`t exist")
	}
	return e, nil
}

// Создаю запись в передаваемом сюда объекте реализующем интерфейс Storage
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

	MakeNewEntry(urlList, uid, longURL)

	return "/" + uid
}

type urlServer struct {
	store *UrlStorage
}

func NewUrlServer() *urlServer {
	store := NewStorageStruct()
	return &urlServer{store: store}
}

func (ts *urlServer) PostHandler(w http.ResponseWriter, req *http.Request) {
	//Для нужной работы конечной точки будем смотреть поля структуры Request
	//Читаем тело запроса- поле Body
	//Поле Body имеет тип io.ReadCloser и данные имеют такой непосредственный вид:
	//&{0xc0001a8000 <nil> <nil> false true {0 0} true false false 0x762820}
	//func io.ReadAll(r io.Reader) ([]byte, error)
	param, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Вроде так нужно
	defer req.Body.Close()

	// Преобразуем тело запроса (тип []byte) в строку:
	longURL := string(param)

	// Генерируем сокращённый URL
	shortURL := req.Host + generateShortURL(ts.store, longURL)

	// Устанавливаем статус ответа 201
	w.WriteHeader(http.StatusCreated)
	// //Content-Type устанавливается как text/plain по умолчанию,
	// //сигнатура такая:
	//w.Header().Set("Content-Type", "text/plain")

	// // Версию HTTP можно узнать так
	// httpVersion := r.Proto

	// Отправляем сокращённый URL в теле ответа
	fmt.Fprint(w, shortURL)

}

func (ts *urlServer) GetHandler(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	longURL, err := ts.store.Get(id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusBadRequest)
		fmt.Println(err)
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
	//
	server := NewUrlServer()

	mux.HandleFunc("POST /{$}", server.PostHandler)
	mux.HandleFunc("GET /{id}", server.GetHandler)
	//mux.Handle("GET /{id}", server.GetHandler)

	http.ListenAndServe("localhost:8080", mux)
}
