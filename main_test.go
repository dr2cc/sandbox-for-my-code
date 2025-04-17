package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUrlStorage_GetHandler(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		input      *UrlStorage
		want       string
		statusCode int
	}{
		{
			name:   "all good",
			method: http.MethodGet,
			input: &UrlStorage{
				Data: map[string]string{"6ba7b811": "https://practicum.yandex.ru/"},
			},
			want:       "https://practicum.yandex.ru/",
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:   "with bad method",
			method: http.MethodPost,
			input: &UrlStorage{
				Data: map[string]string{"6ba7b811": "https://practicum.yandex.ru/"},
			},
			want:       "Method not allowed",
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "key does not match target",
			method: http.MethodGet,
			input: &UrlStorage{
				Data: map[string]string{"6ba7b81": "https://practicum.yandex.ru/"},
			},
			want:       "URL with such id doesn`t exist",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			responseRecorder := httptest.NewRecorder()
			//Так рекомендуют
			request := httptest.NewRequest(tc.method, "/6ba7b811", nil)

			// //Вроде если запрос от клиента, то можно использовать пакет http
			// //но не работает. Оставлю, вдруг что-то подскажет.
			// request, _ := http.NewRequest(tc.method, "/6ba7b811", nil)

			//Вызываем метод GetHandler структуры UrlStorage
			//Этот метод делает запись в responseRecorder
			tc.input.GetHandler(responseRecorder, request)

			// По заданию на конечную точку с методом GET в инкременте 1:
			// В случае успешной обработки запроса сервер возвращает статус с кодом 307
			// и URL (переоеданный ранее) в заголовке "Location"
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Header()["Location"][0]) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}

// Основной каркас сгенерирован VSCode
func TestUrlStorage_PostHandler(t *testing.T) {
	type args struct {
		//w   http.ResponseWriter
		w   *httptest.ResponseRecorder
		req *http.Request
	}

	//Здесь данные не меняющиеся от теста к тесту
	//longURL := "https://practicum.yandex.ru/"
	shortURL := "6ba7b811"
	host := "localhost:8080"
	record := map[string]string{shortURL: "https://practicum.yandex.ru/"}

	tests := []struct {
		name string
		ts   *UrlStorage
		args args
		//
		statusCode int
		want       string
	}{
		{
			name: "all good",
			ts: &UrlStorage{
				Data: record,
			},
			args: args{
				w: httptest.NewRecorder(),
				//req: httptest.NewRequest("POST", "/6ba7b811", nil),
				req: &http.Request{
					Method: "POST",
					Host:   host,
					Body:   io.NopCloser(bytes.NewBuffer([]byte(record[shortURL]))),
				},
			},
			//если все нормально:
			//возвращает статус с кодом 201 (http.StatusCreated)
			//и сокращённым URL в Body (как text/plain)
			statusCode: http.StatusCreated,
			want:       "localhost:8080/6ba7b811",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ts.PostHandler(tt.args.w, tt.args.req)
			//Здесь проверяю статус
			if tt.args.w.Code != tt.statusCode {
				t.Errorf("Want status '%d', got '%d'", tt.statusCode, tt.args.w.Code)
			}
			//Здесь проверяю содержимое Body
			if strings.TrimSpace(tt.args.w.Body.String()) != tt.want {
				t.Errorf("Want '%s', got '%s'", tt.want, tt.args.w.Body)
			}
		})
	}
}
