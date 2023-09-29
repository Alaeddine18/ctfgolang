package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	startPort = 1024
	endPort   = 8192
	timeout   = 2 * time.Second
)

type User struct {
	Username string `json:"User"`
	Secret   string `json:"Secret"`
}

func main() {

	ip, ok := os.LookupEnv("SERVER_IP")
	if !ok {
		log.Fatal("La variable d'environnement SERVER_IP n'est pas définie")
	}

	userName, ok := os.LookupEnv("USER_NAME")
	if !ok {
		log.Fatal("La variable d'environnement USER_NAME n'est pas définie")
	}

	var secret string
	hasher := sha256.New()
	hasher.Write([]byte(userName))
	hashBytes := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)
	secret = hashHex
	println(secret)

	ports := getPorts(ip)
	port := getPing(ip, ports)
	if port == 0 {
		fmt.Println("Aucun port ouvert trouvé")
		return
	}

	postSignUp(ip, port, userName)
	postCheck(ip, port, userName)
	level := postGetUserLevel(ip, port, userName, secret)
	point := postGetUserPoints(ip, port, userName, secret)
	key, key2 := postEnterChallenge(ip, port, userName, secret)

	fmt.Println("Level: ", level)
	fmt.Println("Point: ", point)
	fmt.Println("Key: ", key)
	fmt.Println("Key2: ", key2)

	postSubmitSolution(ip, port, userName, secret, level, point, userName, key2, "http", key)

}

func scanPort(ip string, port int, wg *sync.WaitGroup, c chan int) int {
	defer wg.Done()

	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		return 0
	}

	defer conn.Close()
	fmt.Printf("Port %d est ouvert\n", port)
	c <- port
	return port
}

func getPorts(ip string) []int {
	c := make(chan int, 10)
	var wg sync.WaitGroup

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, timeout)

			if err == nil {
				conn.Close()
				fmt.Printf("Port %d est ouvert\n", port)
				c <- port
			}
		}(port)
	}

	wg.Wait()

	var openPorts []int
	close(c) // Important pour éviter une fuite de goroutine
	for port := range c {
		openPorts = append(openPorts, port)
	}
	return openPorts
}

func getPing(ip string, ports []int) int {
	for _, port := range ports {
		u := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", ip, port),
			Path:   "/ping",
		}
		urlStr := u.String()
		resp, err := http.Get(urlStr)
		if err != nil {
			fmt.Printf("Port %d n'est pas le bon port\n", port)
		} else {
			defer resp.Body.Close()
			return port
		}
	}
	return 0
}

func postSignUp(ip string, port int, userName string) {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
		Path:   "/signup",
	}
	user := User{
		Username: userName,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Erreur lors de la conversion en JSON :", err)
		return
	}
	body := bytes.NewReader(userJSON)
	urlStr := u.String()
	resp, err := http.Post(urlStr, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyparsed, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Réponse du serveur :", string(bodyparsed))

}

func postCheck(ip string, port int, userName string) {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
		Path:   "/check",
	}
	user := User{
		Username: userName,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Erreur lors de la conversion en JSON :", err)
		return
	}
	body := bytes.NewReader(userJSON)
	urlStr := u.String()
	resp, err := http.Post(urlStr, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyparsed, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Réponse du serveur :", string(bodyparsed))

}

func postGetUserLevel(ip string, port int, userName string, secret string) string {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
		Path:   "/getUserLevel",
	}
	user := User{
		Username: userName,
		Secret:   secret,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Erreur lors de la conversion en JSON :", err)
		return ""
	}
	body := bytes.NewReader(userJSON)
	urlStr := u.String()
	resp, err := http.Post(urlStr, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	levelParsed, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := strings.TrimPrefix(string(levelParsed), "Level: ")
	result = strings.ReplaceAll(result, "\n", "")

	return result

}

func postGetUserPoints(ip string, port int, userName string, secret string) string {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
		Path:   "/getUserPoints",
	}
	user := User{
		Username: userName,
		Secret:   secret,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Erreur lors de la conversion en JSON :", err)
		return ""
	}
	body := bytes.NewReader(userJSON)
	urlStr := u.String()
	resp, err := http.Post(urlStr, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	pointParsed, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := strings.TrimPrefix(string(pointParsed), "User points: "+userName)
	result = strings.ReplaceAll(result, "\n", "")

	return result

}

func postEnterChallenge(ip string, port int, userName string, secret string) (string, string) {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
		Path:   "/enterChallenge",
	}
	user := User{
		Username: userName,
		Secret:   secret,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Erreur lors de la conversion en JSON :", err)
		return "", ""
	}
	body := bytes.NewReader(userJSON)
	urlStr := u.String()
	resp, err := http.Post(urlStr, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	key, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	startIndex := 57
	endFirstKeyIndex := 108

	firstKey := key[startIndex:89]
	secondKey := key[endFirstKeyIndex:]

	var parsedKey string = string(firstKey)
	parsedKeyWithoutN := strings.ReplaceAll(parsedKey, "\n", "")

	var parsedSecondKey string = string(secondKey)
	parsedKeyWithoutNKey := strings.ReplaceAll(parsedSecondKey, "\n", "")

	return parsedKeyWithoutN, parsedKeyWithoutNKey

}

func postSubmitSolution(ip string, port int, userName string, secret string, level string, points string, challengeUsername string, challengeSecret string, protocol string, secretKey string) {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
		Path:   "/submitSolution",
	}

	// Créez la structure JSON selon le format souhaité
	userJSON, err := json.Marshal(struct {
		User    string `json:"User"`
		Secret  string `json:"Secret"`
		Content struct {
			Level     string `json:"Level"`
			Challenge struct {
				Username string `json:"Username"`
				Secret   string `json:"Secret"`
				Points   string `json:"Points"`
			} `json:"Challenge"`
			Protocol  string `json:"Protocol"`
			SecretKey string `json:"SecretKey"`
		} `json:"Content"`
	}{
		User:   userName,
		Secret: secret,
		Content: struct {
			Level     string `json:"Level"`
			Challenge struct {
				Username string `json:"Username"`
				Secret   string `json:"Secret"`
				Points   string `json:"Points"`
			} `json:"Challenge"`
			Protocol  string `json:"Protocol"`
			SecretKey string `json:"SecretKey"`
		}{
			Level: level,
			Challenge: struct {
				Username string `json:"Username"`
				Secret   string `json:"Secret"`
				Points   string `json:"Points"`
			}{
				Username: challengeUsername,
				Secret:   challengeSecret,
				Points:   points,
			},
			Protocol:  protocol,
			SecretKey: secretKey,
		},
	})

	if err != nil {
		fmt.Println("Erreur lors de la conversion en JSON :", err)
		return
	}

	body := bytes.NewReader(userJSON)
	urlStr := u.String()
	resp, err := http.Post(urlStr, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Réponse du serveur :", string(response))
}
