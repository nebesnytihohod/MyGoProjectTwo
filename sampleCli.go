package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-gcfg/gcfg"
	"github.com/joho/godotenv"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
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
	logFile, _ := os.Create("./log.txt")
	defer logFile.Close()

	logger := log.New(logFile, "[myAppLog]: ", log.LstdFlags|log.Lshortfile)

	fmt.Println("Outside a goroutine.")
	log.Println("[LOG] Outside a goroutine.") // Вариант с простым логгером для сравнения
	logger.Println("Outside a goroutine.")    // Вариант с продвинутым логгером для сравнения

	go func() { //Обьявление анонимной функции ии вызов ее как сопрограммы
		fmt.Println("Inside a goroutine.")
		log.Println("[LOG] Inside a goroutine.")
		logger.Println("Inside a goroutine.")
	}()

	fmt.Println("Outside a goroutine again.")
	log.Println("[LOG] Outside a goroutine again.")
	logger.Println("Outside a goroutine again.")

	// конфиг в JSON - начало
	configFileJson, _ := os.Open("config.json")
	defer configFileJson.Close()

	decoder := json.NewDecoder(configFileJson)
	appConf1 := appConfiguration1{}
	err := decoder.Decode(&appConf1)
	if err != nil {
		fmt.Println("Error in config JSON:", err)
	}
	// конфиг в JSON - конец
	fmt.Println("[JSON] Enabled:", appConf1.Enabled)
	fmt.Println("[JSON] Path:", appConf1.Path)

	// конфиг в YAML - начало
	configFileYaml, err := yaml.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("Error in config YAML:", err)
	}
	enabled, _ := configFileYaml.GetBool("enable")
	path, _ := configFileYaml.Get("path")
	// конфиг в YAML - конец
	fmt.Println("[YAML] Enabled:", enabled)
	fmt.Println("[YAML] Path", path)

	// конфиг в INI - начало
	appConf2 := appConfiguration2{}
	err = gcfg.ReadFileInto(&appConf2, "config.ini")
	if err != nil {
		fmt.Println("Error in config INI:", err)
	}
	// конфиг в INI - конец
	fmt.Println("[INI] Enabled:", appConf2.Section1.Enabled)
	fmt.Println("[INI] Path:", appConf2.Section1.Path)
	fmt.Println("[INI] Name:", appConf2.Section2.Name)
	fmt.Println("[INI] Surame:", appConf2.Section2.Surname)

	runtime.Gosched() // Обращение к планировщику

	var waitGroup sync.WaitGroup // Обьявлени переменной счетчика группы ожидания
	var i int = -1
	var cliArg string

	fmt.Println(os.Args)

	for i, cliArg = range os.Args[1:] { // Примитивный способ считывания аргументов вызова программы
		fmt.Println(i, " "+cliArg)
		waitGroup.Add(1)      // Добавляем к счетчику 1 для подсчета числа запускаемых сопрограмм
		go func(cla string) { // Некая функция чтобы что-то сделать полезное
			fmt.Println(cla) // Тут что-то делаем
			waitGroup.Done() // Сообщаем, что выполнение сопрограммы завершено когда функция заканчивает свою работу
		}(cliArg)
	}

	waitGroup.Wait() // Внешняя сопрограмма ожидает пока все вызванные внутри нее сопрограммы заверщат свою работу (вызовут waitGroup.Done(), и счетчик запусков обнулится

	fmt.Printf("Printed %d cli arguments\n", i+1)

	// Использование разных вариантов web-сервера
	//http.HandleFunc("/", hello) //регистация, создание обработчика пути
	//http.ListenAndServe("localhost:4000", nil)
	fmt.Println("port:", os.Getenv("MYPORT")) // использование переменных среды ОС
	//http.ListenAndServe("localhost"+":"+os.Getenv("MYPORT"), nil)
	//http.ListenAndServe(":"+os.Getenv("MYPORT"), nil) // запуск сервера

	//errEnv := godotenv.Load() // load .env-file from default path
	errEnv := godotenv.Load("specenv.env") // load enf-file from specified path
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	secretKey := os.Getenv("SECRET_KEY")

	fmt.Println("S3 bucket: ", s3Bucket)
	fmt.Println("secret key: ", secretKey)

	pathRes := newPathResolver()                                 // получение экземпляра маршрутизатора
	pathRes.Add("GET /hello", hello)                             // регистрация пути к endpoint и его функции
	pathRes.Add("(GET|HEAD) /goodbye(/?[A-Za-z0-9]*)?", goodbye) // регистрация пути к endpoint и его функции
	http.ListenAndServe(":"+os.Getenv("MYPORT"), pathRes)        // запуск сервера

}
