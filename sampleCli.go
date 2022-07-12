package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-gcfg/gcfg"
	"github.com/kylelemons/go-gypsy/yaml"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type appConfiguration1 struct {
	Enabled bool
	Path    string
}

type appConfiguration2 struct {
	Section1 struct {
		Enabled bool
		Path    string
	}
	Section2 struct {
		Name    string
		Surname string
	}
}

//func hello(res http.ResponseWriter, req *http.Request) {
//	fmt.Fprint(res, "Hello, World and Max!")
//}

// функции для работы расширенного сервера
// описание структуры содержащей функции обработки endpoints и регулярных выражений для путей
type regexResolver struct {
	handlers map[string]http.HandlerFunc
	cache    map[string]*regexp.Regexp
}

// описание метода для добавления-регистрации путей и их функций
func (r *regexResolver) Add(regex string, handler http.HandlerFunc) {
	r.handlers[regex] = handler
	cache, _ := regexp.Compile(regex)
	r.cache[regex] = cache
}

// описание метода для поиска и вызова функции-обработчика
func (r *regexResolver) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	check := req.Method + " " + req.URL.Path
	for pattern, handlerFunc := range r.handlers {
		if r.cache[pattern].MatchString(check) == true {
			handlerFunc(res, req)
			return
		}
	}
	http.NotFound(res, req) // если соответствие не найдено вернуть ошибку 404-"Page Not Found"
}

// описание обработчика путей-маршрутов
func newPathResolver() *regexResolver {
	return &regexResolver{
		handlers: make(map[string]http.HandlerFunc),
		cache:    make(map[string]*regexp.Regexp),
	}
}

// функция для обработки endpoint - hello
func hello(res http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()  // разбор строки запроса
	name := query.Get("name") // получение значения name из параметра запроса /hello?name=
	if name == "" {
		name = "Maxim"
	}
	fmt.Fprint(res, "Hello, my name is ", name, "\n")
	fmt.Fprint(res, "\n[LOG]Query: ", query)
}

// функция для обработки endpoint - goodbye
func goodbye(res http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(path, "/")
	name := ""
	if len(parts) > 2 {
		name = parts[2]
	}
	if name == "" {
		name = "Maxim"
	}
	fmt.Fprint(res, "Goodbye,", name, "\n")
	fmt.Fprint(res, "\n[LOG]URL path: ", path, "\n")
}

func main() {
	/*	app := cli.NewApp()
		app.Usage = "Count up or down."
		app.Commands = []cli.Command{ // Определение одной или нескольких команд
			{
				Name: "up", ShortName: "u",
				Usage: "Count Up",
				// Определение параметров команды
				Flags: []cli.Flag{
					cli.IntFlag{
						Name: "stop, s",
						Usage: "Value to count up to",
						Value: 10,
					},
				},
			},
			Action: func(c *cli.Context) error {
				start := c.Int("stop") // Получение параметра команды

				return nil
			},
		},
		{
			Name:"down", ShortName: "d",
			Usage: "Count Down",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "start, s",
					Usage: "Start counting down from",
					Value: 10,
				},
			},
		},
			Action: func(c *cli.Context) error {
			start := c.Int("start")

			}
			return nil
			},
		},

		app.Run(os.Args)*/

	configFileJson, _ := os.Open("config.json")
	defer configFileJson.Close()

	decoder := json.NewDecoder(configFileJson)
	appConf1 := appConfiguration1{}
	err := decoder.Decode(&appConf1)
	if err != nil {
		fmt.Println("Error in config JSON:", err)
	}

	fmt.Println("[JSON] Enabled:", appConf1.Enabled)
	fmt.Println("[JSON] Path:", appConf1.Path)

	configFileYaml, err := yaml.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("Error in config YAML:", err)
	}
	enabled, _ := configFileYaml.GetBool("enable")
	path, _ := configFileYaml.Get("path")
	fmt.Println("[YAML] Enabled:", enabled)
	fmt.Println("[YAML] Path", path)

	appConf2 := appConfiguration2{}
	err = gcfg.ReadFileInto(&appConf2, "config.ini")
	if err != nil {
		fmt.Println("Error in config INI:", err)
	}
	fmt.Println("[INI] Enabled:", appConf2.Section1.Enabled)
	fmt.Println("[INI] Path:", appConf2.Section1.Path)
	fmt.Println("[INI] Name:", appConf2.Section2.Name)
	fmt.Println("[INI] Surame:", appConf2.Section2.Surname)

	//http.HandleFunc("/", hello) //регистация, создание обработчика пути
	//http.ListenAndServe("localhost:4000", nil)
	fmt.Println("port:", os.Getenv("MYPORT"))
	//http.ListenAndServe("localhost"+":"+os.Getenv("MYPORT"), nil)
	//http.ListenAndServe(":"+os.Getenv("MYPORT"), nil) // запуск сервера

	pathRes := newPathResolver()                                 // получение экземпляра маршрутизатора
	pathRes.Add("GET /hello", hello)                             // регистрация пути к endpoint и его функции
	pathRes.Add("(GET|HEAD) /goodbye(/?[A-Za-z0-9]*)?", goodbye) // регистрация пути к endpoint и его функции
	http.ListenAndServe(":"+os.Getenv("MYPORT"), pathRes)        // запуск сервера

}
