package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/satori/go.uuid"
    "gopkg.in/yaml.v2"
)

var workers = -1
var shutdown = false
var shutdownKey string
var shutdownToken string

type TwilioConfiguration struct {
    AccountSid string `yaml:"accountSid"`
    AuthToken string `yaml:"authToken"`
    FromNumber string `yaml:"fromNumber"`
    ToNumber string `yaml:"toNumber"`
}

type WatchConfiguration struct {
    Twilio TwilioConfiguration `yaml:"twilio"`
    NiceHashAddr string `yaml:"niceHashAddr"`
    ShutdownSecret string `yaml:"shutdownSecret"`
}

type NiceHashStatsResult struct {
    Address string        `json:"addr"`
    Workers []interface{} `json:"workers"`
}

type NiceHashStats struct {
    Result NiceHashStatsResult `json:"result"`
}

var config WatchConfiguration

func notify(message string) {
    urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + config.Twilio.AccountSid + "/Messages.json"

    msgData := url.Values{}
    msgData.Set("To", config.Twilio.ToNumber)
    msgData.Set("From", config.Twilio.FromNumber)
    msgData.Set("Body", message)
    msgDataReader := *strings.NewReader(msgData.Encode())

    client := &http.Client{}
    req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
    req.SetBasicAuth(config.Twilio.AccountSid, config.Twilio.AuthToken)
    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    resp, _ := client.Do(req)
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        var data map[string]interface{}
        decoder := json.NewDecoder(resp.Body)
        err := decoder.Decode(&data)
        if err != nil {
            fmt.Println(err)
        }
    } else {
        fmt.Println(resp.Status)
    }
}

func watch(done chan string) {
    for shutdown == false {
        fmt.Println("Polling...")

        response, err := http.Get("https://api.nicehash.com/api?method=stats.provider.workers&addr=" + config.NiceHashAddr)
        if err != nil {
            fmt.Printf("The HTTP request failed with error %s\n", err)
        } else {
            data, _ := ioutil.ReadAll(response.Body)
            var stats NiceHashStats
            err := json.Unmarshal([]byte(data), &stats)
            if err != nil {
                fmt.Println(err)
                fmt.Printf("%+v\n", stats)
            } else {
                // Number of workers can change based on what is currently being mined
                // For now, just let me know if we go to or from 0 workers.
                if workers != -1 && workers != len(stats.Result.Workers) && (workers == 0 || len(stats.Result.Workers) == 0) {
                    fmt.Println("Found change, sending text.")
                    notify("NiceHash Workers: "+strconv.Itoa(len(stats.Result.Workers))+", was: "+strconv.Itoa(workers))
                }
                workers = len(stats.Result.Workers)
                fmt.Printf("%d workers\n", len(stats.Result.Workers))
            }
        }

        time.Sleep(5 * time.Minute)
    }

    notify("NiceHash.Watcher Shutting Down...")

    done <- "My watch has ended."
}

func main() {
	fmt.Println("Reading configuration...")

	configData, err := ioutil.ReadFile("config.yml")
    if err != nil {
        panic(err)
    }

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
	    panic(err)
	}

	fmt.Print(config.NiceHashAddr)

    shutdownKey = uuid.NewV4().String()

    sig := hmac.New(sha256.New, []byte(config.ShutdownSecret))
    sig.Write([]byte(shutdownKey))
    shutdownToken = hex.EncodeToString(sig.Sum(nil))
    fmt.Println(shutdownToken)

    fmt.Println("Starting the application...")

    gin.SetMode(gin.ReleaseMode)

    router := gin.Default()
    router.LoadHTMLGlob("templates/*")

    router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.tmpl", gin.H{
            "workers": workers,
        })
    })

    router.GET("/shutdown/key", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "key": shutdownKey,
        })
    })

    router.POST("/shutdown", func(c *gin.Context) {
        token := c.Request.FormValue("api_token")

        if token == shutdownToken {
            c.JSON(200, gin.H{
                "message": "ok",
            })

            shutdown = true
        } else {
            c.JSON(200, gin.H{
                "message": "no",
            })
        }
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    watch_done := make(chan string)
    go watch(watch_done)
    go router.Run(":" + port)

    _ = <-watch_done

    fmt.Println("Terminating the application...")
}
