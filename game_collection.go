// game_collection.go
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Genre     string `json:"genre"`
	Platform  string `json:"platform"`
	Year      int    `json:"year"`
	Completed bool   `json:"completed"`
	Rating    int    `json:"rating"`
	Notes     string `json:"notes"`
	AddedDate string `json:"added_date"`
}

type GamesData struct {
	Games []Game `json:"games"`
}

type GameCollection struct {
	games  []Game
	nextID int
}

func NewGameCollection() *GameCollection {
	return &GameCollection{
		games:  []Game{},
		nextID: 1,
	}
}

func (c *GameCollection) AddGame(title, genre, platform string, year int, completed bool, rating int, notes string) (Game, error) {
	if rating < 1 || rating > 10 {
		return Game{}, fmt.Errorf("оценка должна быть от 1 до 10")
	}
	currentYear := time.Now().Year()
	if year < 1980 || year > currentYear {
		return Game{}, fmt.Errorf("год должен быть от 1980 до %d", currentYear)
	}
	if title == "" || genre == "" || platform == "" {
		return Game{}, fmt.Errorf("название, жанр и платформа не могут быть пустыми")
	}
	game := Game{
		ID:        c.nextID,
		Title:     title,
		Genre:     genre,
		Platform:  platform,
		Year:      year,
		Completed: completed,
		Rating:    rating,
		Notes:     notes,
		AddedDate: time.Now().Format("2006-01-02"),
	}
	c.games = append(c.games, game)
	c.nextID++
	return game, nil
}

func (c *GameCollection) FindGame(id int) *Game {
	for i := range c.games {
		if c.games[i].ID == id {
			return &c.games[i]
		}
	}
	return nil
}

func (c *GameCollection) EditGame(id int, updates map[string]interface{}) bool {
	game := c.FindGame(id)
	if game == nil {
		return false
	}
	for key, value := range updates {
		switch key {
		case "title":
			if v, ok := value.(string); ok {
				game.Title = v
			}
		case "genre":
			if v, ok := value.(string); ok {
				game.Genre = v
			}
		case "platform":
			if v, ok := value.(string); ok {
				game.Platform = v
			}
		case "year":
			if v, ok := value.(int); ok {
				game.Year = v
			}
		case "completed":
			if v, ok := value.(bool); ok {
				game.Completed = v
			}
		case "rating":
			if v, ok := value.(int); ok {
				game.Rating = v
			}
		case "notes":
			if v, ok := value.(string); ok {
				game.Notes = v
			}
		}
	}
	return true
}

func (c *GameCollection) DeleteGame(id int) bool {
	for i, g := range c.games {
		if g.ID == id {
			c.games = append(c.games[:i], c.games[i+1:]...)
			return true
		}
	}
	return false
}

func (c *GameCollection) SearchGames(query string) []Game {
	q := strings.ToLower(query)
	var result []Game
	for _, g := range c.games {
		if strings.Contains(strings.ToLower(g.Title), q) {
			result = append(result, g)
		}
	}
	return result
}

func (c *GameCollection) FilterByCompleted(completed bool) []Game {
	var result []Game
	for _, g := range c.games {
		if g.Completed == completed {
			result = append(result, g)
		}
	}
	return result
}

func (c *GameCollection) FilterByGenre(genre string) []Game {
	var result []Game
	for _, g := range c.games {
		if strings.EqualFold(g.Genre, genre) {
			result = append(result, g)
		}
	}
	return result
}

func (c *GameCollection) FilterByPlatform(platform string) []Game {
	var result []Game
	for _, g := range c.games {
		if strings.EqualFold(g.Platform, platform) {
			result = append(result, g)
		}
	}
	return result
}

func (c *GameCollection) SortByRating(reverse bool) []Game {
	sorted := make([]Game, len(c.games))
	copy(sorted, c.games)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if reverse {
				if sorted[i].Rating < sorted[j].Rating {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			} else {
				if sorted[i].Rating > sorted[j].Rating {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}
	return sorted
}

func (c *GameCollection) SortByTitle() []Game {
	sorted := make([]Game, len(c.games))
	copy(sorted, c.games)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if strings.ToLower(sorted[i].Title) > strings.ToLower(sorted[j].Title) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func (c *GameCollection) GetStats() map[string]interface{} {
	total := len(c.games)
	completed := len(c.FilterByCompleted(true))
	uncompleted := total - completed
	var ratings []int
	for _, g := range c.games {
		if g.Completed {
			ratings = append(ratings, g.Rating)
		}
	}
	avgRating := 0.0
	if len(ratings) > 0 {
		sum := 0
		for _, r := range ratings {
			sum += r
		}
		avgRating = float64(sum) / float64(len(ratings))
	}
	platforms := make(map[string]int)
	genres := make(map[string]int)
	for _, g := range c.games {
		platforms[g.Platform]++
		genres[g.Genre]++
	}
	return map[string]interface{}{
		"total":       total,
		"completed":   completed,
		"uncompleted": uncompleted,
		"avg_rating":  avgRating,
		"platforms":   platforms,
		"genres":      genres,
	}
}

func (c *GameCollection) SaveToFile(filename string) error {
	data := GamesData{Games: c.games}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (c *GameCollection) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var gd GamesData
	if err := json.Unmarshal(data, &gd); err != nil {
		return err
	}
	c.games = gd.Games
	for _, g := range c.games {
		if g.ID >= c.nextID {
			c.nextID = g.ID + 1
		}
	}
	return nil
}

func (c *GameCollection) ExportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()
	headers := []string{"ID", "Название", "Жанр", "Платформа", "Год", "Пройдена", "Оценка", "Заметки", "Дата добавления"}
	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, g := range c.games {
		completedStr := "Нет"
		if g.Completed {
			completedStr = "Да"
		}
		row := []string{
			strconv.Itoa(g.ID),
			g.Title,
			g.Genre,
			g.Platform,
			strconv.Itoa(g.Year),
			completedStr,
			strconv.Itoa(g.Rating),
			g.Notes,
			g.AddedDate,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (c *GameCollection) ImportCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("файл пуст или нет данных")
	}
	headers := records[0]
	// Проверяем заголовки (упрощённо)
	for _, row := range records[1:] {
		if len(row) < 9 {
			continue
		}
		title := row[1]
		genre := row[2]
		platform := row[3]
		year, _ := strconv.Atoi(row[4])
		completed := row[5] == "Да"
		rating, _ := strconv.Atoi(row[6])
		notes := row[7]
		_, err := c.AddGame(title, genre, platform, year, completed, rating, notes)
		if err != nil {
			fmt.Println("Ошибка импорта строки:", err)
		}
	}
	return nil
}

func readString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readInt(prompt string) int {
	for {
		input := readString(prompt)
		if val, err := strconv.Atoi(input); err == nil {
			return val
		}
		fmt.Println("Введите число.")
	}
}

func readBool(prompt string) bool {
	for {
		input := readString(prompt)
		if input == "1" {
			return true
		} else if input == "0" {
			return false
		}
		fmt.Println("Введите 1 или 0.")
	}
}

func printGame(game Game) {
	status := "✅ Пройдена"
	if !game.Completed {
		status = "⏳ Не пройдена"
	}
	fmt.Printf("#%d - %s (%d)\n", game.ID, game.Title, game.Year)
	fmt.Printf("   Жанр: %s, Платформа: %s\n", game.Genre, game.Platform)
	fmt.Printf("   %s, Оценка: %d/10\n", status, game.Rating)
	if game.Notes != "" {
		fmt.Printf("   Заметки: %s\n", game.Notes)
	}
	fmt.Printf("   Добавлена: %s\n", game.AddedDate)
}

func main() {
	collection := NewGameCollection()
	if err := collection.LoadFromFile("games_data.json"); err != nil {
		fmt.Println("Ошибка загрузки:", err)
	}

	for {
		fmt.Println("\n===== КОЛЛЕКЦИЯ ИГР (Go) =====")
		fmt.Println("1. Добавить игру")
		fmt.Println("2. Показать все игры")
		fmt.Println("3. Показать пройденные игры")
		fmt.Println("4. Показать непройденные игры")
		fmt.Println("5. Найти игры по названию")
		fmt.Println("6. Сортировать по оценке (по убыванию)")
		fmt.Println("7. Сортировать по названию")
		fmt.Println("8. Редактировать игру")
		fmt.Println("9. Удалить игру")
		fmt.Println("10. Показать статистику")
		fmt.Println("11. Сохранить в файл")
		fmt.Println("12. Загрузить из файла")
		fmt.Println("13. Экспорт в CSV")
		fmt.Println("14. Импорт из CSV")
		fmt.Println("0. Выход")
		choice := readString("Выберите действие: ")

		switch choice {
		case "0":
			return
		case "1":
			title := readString("Название: ")
			if title == "" {
				fmt.Println("Название не может быть пустым.")
				continue
			}
			genre := readString("Жанр: ")
			if genre == "" {
				fmt.Println("Жанр не может быть пустым.")
				continue
			}
			platform := readString("Платформа: ")
			if platform == "" {
				fmt.Println("Платформа не может быть пустой.")
				continue
			}
			year := readInt("Год выпуска: ")
			completed := readBool("Статус (1-пройдена, 0-нет): ")
			rating := readInt("Оценка (1-10): ")
			notes := readString("Заметки (необязательно): ")
			game, err := collection.AddGame(title, genre, platform, year, completed, rating, notes)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Printf("Игра добавлена с ID %d\n", game.ID)
			}
		case "2":
			if len(collection.games) == 0 {
				fmt.Println("Нет игр.")
			} else {
				for _, g := range collection.games {
					printGame(g)
				}
			}
		case "3":
			games := collection.FilterByCompleted(true)
			if len(games) == 0 {
				fmt.Println("Нет пройденных игр.")
			} else {
				for _, g := range games {
					printGame(g)
				}
			}
		case "4":
			games := collection.FilterByCompleted(false)
			if len(games) == 0 {
				fmt.Println("Нет непройденных игр.")
			} else {
				for _, g := range games {
					printGame(g)
				}
			}
		case "5":
			query := readString("Введите часть названия: ")
			results := collection.SearchGames(query)
			if len(results) == 0 {
				fmt.Println("Игры не найдены.")
			} else {
				for _, g := range results {
					printGame(g)
				}
			}
		case "6":
			sorted := collection.SortByRating(true)
			if len(sorted) == 0 {
				fmt.Println("Нет игр.")
			} else {
				for _, g := range sorted {
					printGame(g)
				}
			}
		case "7":
			sorted := collection.SortByTitle()
			if len(sorted) == 0 {
				fmt.Println("Нет игр.")
			} else {
				for _, g := range sorted {
					printGame(g)
				}
			}
		case "8":
			id := readInt("Введите ID игры для редактирования: ")
			game := collection.FindGame(id)
			if game == nil {
				fmt.Println("Игра не найдена.")
				continue
			}
			fmt.Println("Оставьте поле пустым, чтобы не менять.")
			newTitle := readString(fmt.Sprintf("Название (%s): ", game.Title))
			newGenre := readString(fmt.Sprintf("Жанр (%s): ", game.Genre))
			newPlatform := readString(fmt.Sprintf("Платформа (%s): ", game.Platform))
			newYear := readString(fmt.Sprintf("Год (%d): ", game.Year))
			newCompleted := readString(fmt.Sprintf("Статус (1-пройдена, 0-нет) сейчас: %d: ", map[bool]int{true: 1, false: 0}[game.Completed]))
			newRating := readString(fmt.Sprintf("Оценка (%d): ", game.Rating))
			newNotes := readString(fmt.Sprintf("Заметки (%s): ", game.Notes))
			updates := make(map[string]interface{})
			if newTitle != "" {
				updates["title"] = newTitle
			}
			if newGenre != "" {
				updates["genre"] = newGenre
			}
			if newPlatform != "" {
				updates["platform"] = newPlatform
			}
			if newYear != "" {
				if y, err := strconv.Atoi(newYear); err == nil {
					updates["year"] = y
				} else {
					fmt.Println("Год должен быть числом, пропускаем.")
				}
			}
			if newCompleted != "" {
				updates["completed"] = newCompleted == "1"
			}
			if newRating != "" {
				if r, err := strconv.Atoi(newRating); err == nil {
					updates["rating"] = r
				} else {
					fmt.Println("Оценка должна быть числом, пропускаем.")
				}
			}
			if newNotes != "" {
				updates["notes"] = newNotes
			}
			if collection.EditGame(id, updates) {
				fmt.Println("Игра обновлена.")
			} else {
				fmt.Println("Ошибка обновления.")
			}
		case "9":
			id := readInt("Введите ID игры для удаления: ")
			if collection.DeleteGame(id) {
				fmt.Println("Игра удалена.")
			} else {
				fmt.Println("Игра не найдена.")
			}
		case "10":
			stats := collection.GetStats()
			fmt.Println("\n=== СТАТИСТИКА ===")
			fmt.Printf("Всего игр: %d\n", stats["total"])
			fmt.Printf("Пройдено: %d\n", stats["completed"])
			fmt.Printf("Не пройдено: %d\n", stats["uncompleted"])
			fmt.Printf("Средняя оценка (только пройденные): %.2f\n", stats["avg_rating"])
			fmt.Println("По платформам:")
			platforms := stats["platforms"].(map[string]int)
			for p, c := range platforms {
				fmt.Printf("  %s: %d\n", p, c)
			}
			fmt.Println("По жанрам:")
			genres := stats["genres"].(map[string]int)
			for g, c := range genres {
				fmt.Printf("  %s: %d\n", g, c)
			}
		case "11":
			if err := collection.SaveToFile("games_data.json"); err != nil {
				fmt.Println("Ошибка сохранения:", err)
			} else {
				fmt.Println("Сохранено.")
			}
		case "12":
			if err := collection.LoadFromFile("games_data.json"); err != nil {
				fmt.Println("Ошибка загрузки:", err)
			} else {
				fmt.Println("Загружено.")
			}
		case "13":
			if err := collection.ExportCSV("games_export.csv"); err != nil {
				fmt.Println("Ошибка экспорта:", err)
			} else {
				fmt.Println("Экспортировано в games_export.csv")
			}
		case "14":
			if err := collection.ImportCSV("games_export.csv"); err != nil {
				fmt.Println("Ошибка импорта:", err)
			} else {
				fmt.Println("Импортировано из games_export.csv")
			}
		default:
			fmt.Println("Неизвестная команда.")
		}
	}
}
