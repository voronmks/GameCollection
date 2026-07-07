// GameCollection.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

public record Game(
    int Id,
    string Title,
    string Genre,
    string Platform,
    int Year,
    bool Completed,
    int Rating,
    string Notes,
    string AddedDate
);

public class GamesData
{
    public List<Game> Games { get; set; } = new();
}

public class GameCollection
{
    private List<Game> games = new();
    private int nextId = 1;

    public IReadOnlyList<Game> Games => games.AsReadOnly();

    public Game AddGame(string title, string genre, string platform, int year, bool completed, int rating, string notes)
    {
        if (rating < 1 || rating > 10) throw new ArgumentException("Оценка должна быть от 1 до 10");
        if (year < 1980 || year > DateTime.Now.Year)
            throw new ArgumentException($"Год должен быть от 1980 до {DateTime.Now.Year}");
        if (string.IsNullOrWhiteSpace(title) || string.IsNullOrWhiteSpace(genre) || string.IsNullOrWhiteSpace(platform))
            throw new ArgumentException("Название, жанр и платформа не могут быть пустыми");
        if (notes == null) notes = "";
        var game = new Game(nextId, title, genre, platform, year, completed, rating, notes, DateTime.Now.ToString("yyyy-MM-dd"));
        games.Add(game);
        nextId++;
        return game;
    }

    public Game? FindGame(int id) => games.FirstOrDefault(g => g.Id == id);

    public bool EditGame(int id, Dictionary<string, object> updates)
    {
        var old = FindGame(id);
        if (old == null) return false;
        games.Remove(old);
        string title = updates.ContainsKey("title") ? (string)updates["title"] : old.Title;
        string genre = updates.ContainsKey("genre") ? (string)updates["genre"] : old.Genre;
        string platform = updates.ContainsKey("platform") ? (string)updates["platform"] : old.Platform;
        int year = updates.ContainsKey("year") ? (int)updates["year"] : old.Year;
        bool completed = updates.ContainsKey("completed") ? (bool)updates["completed"] : old.Completed;
        int rating = updates.ContainsKey("rating") ? (int)updates["rating"] : old.Rating;
        string notes = updates.ContainsKey("notes") ? (string)updates["notes"] : old.Notes;
        var updated = new Game(old.Id, title, genre, platform, year, completed, rating, notes, old.AddedDate);
        games.Add(updated);
        return true;
    }

    public bool DeleteGame(int id) => games.RemoveAll(g => g.Id == id) > 0;

    public List<Game> SearchGames(string query)
    {
        var q = query.ToLower();
        return games.Where(g => g.Title.ToLower().Contains(q)).ToList();
    }

    public List<Game> FilterByCompleted(bool completed) => games.Where(g => g.Completed == completed).ToList();

    public List<Game> FilterByGenre(string genre) =>
        games.Where(g => string.Equals(g.Genre, genre, StringComparison.OrdinalIgnoreCase)).ToList();

    public List<Game> FilterByPlatform(string platform) =>
        games.Where(g => string.Equals(g.Platform, platform, StringComparison.OrdinalIgnoreCase)).ToList();

    public List<Game> SortByRating(bool reverse) =>
        games.OrderBy(g => g.Rating).ToList(); // потом можно развернуть

    public List<Game> SortByTitle() =>
        games.OrderBy(g => g.Title, StringComparer.OrdinalIgnoreCase).ToList();

    public Dictionary<string, object> GetStats()
    {
        int total = games.Count;
        int completedCount = FilterByCompleted(true).Count;
        int uncompleted = total - completedCount;
        double avgRating = FilterByCompleted(true).Any() ? FilterByCompleted(true).Average(g => g.Rating) : 0;
        var platforms = games.GroupBy(g => g.Platform).ToDictionary(g => g.Key, g => g.Count());
        var genres = games.GroupBy(g => g.Genre).ToDictionary(g => g.Key, g => g.Count());
        return new Dictionary<string, object>
        {
            ["total"] = total,
            ["completed"] = completedCount,
            ["uncompleted"] = uncompleted,
            ["avg_rating"] = avgRating,
            ["platforms"] = platforms,
            ["genres"] = genres
        };
    }

    public void SaveToFile(string filename)
    {
        var data = new GamesData { Games = games };
        var options = new JsonSerializerOptions { WriteIndented = true };
        string json = JsonSerializer.Serialize(data, options);
        File.WriteAllText(filename, json);
    }

    public void LoadFromFile(string filename)
    {
        if (!File.Exists(filename)) return;
        string json = File.ReadAllText(filename);
        var data = JsonSerializer.Deserialize<GamesData>(json);
        if (data != null)
        {
            games = data.Games;
            nextId = games.Any() ? games.Max(g => g.Id) + 1 : 1;
        }
    }

    public void ExportCSV(string filename)
    {
        using var writer = new StreamWriter(filename);
        writer.WriteLine("ID;Название;Жанр;Платформа;Год;Пройдена;Оценка;Заметки;Дата добавления");
        foreach (var g in games)
        {
            writer.WriteLine($"{g.Id};{g.Title};{g.Genre};{g.Platform};{g.Year};{(g.Completed ? "Да" : "Нет")};{g.Rating};{g.Notes};{g.AddedDate}");
        }
    }

    public void ImportCSV(string filename)
    {
        if (!File.Exists(filename)) throw new FileNotFoundException("Файл не найден");
        using var reader = new StreamReader(filename);
        string header = reader.ReadLine(); // skip header
        while (!reader.EndOfStream)
        {
            string line = reader.ReadLine();
            var parts = line.Split(';');
            if (parts.Length < 9) continue;
            string title = parts[1];
            string genre = parts[2];
            string platform = parts[3];
            int year = int.Parse(parts[4]);
            bool completed = parts[5] == "Да";
            int rating = int.Parse(parts[6]);
            string notes = parts[7];
            try
            {
                AddGame(title, genre, platform, year, completed, rating, notes);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Ошибка импорта строки: {ex.Message}");
            }
        }
    }
}

public static class Program
{
    private static string ReadString(string prompt)
    {
        Console.Write(prompt);
        return Console.ReadLine()?.Trim() ?? "";
    }

    private static int ReadInt(string prompt)
    {
        while (true)
        {
            Console.Write(prompt);
            if (int.TryParse(Console.ReadLine(), out int result))
                return result;
            Console.WriteLine("Введите число.");
        }
    }

    private static bool ReadBool(string prompt)
    {
        while (true)
        {
            string input = ReadString(prompt);
            if (input == "1") return true;
            if (input == "0") return false;
            Console.WriteLine("Введите 1 или 0.");
        }
    }

    private static void PrintGame(Game game)
    {
        string status = game.Completed ? "✅ Пройдена" : "⏳ Не пройдена";
        Console.WriteLine($"#{game.Id} - {game.Title} ({game.Year})");
        Console.WriteLine($"   Жанр: {game.Genre}, Платформа: {game.Platform}");
        Console.WriteLine($"   {status}, Оценка: {game.Rating}/10");
        if (!string.IsNullOrWhiteSpace(game.Notes))
            Console.WriteLine($"   Заметки: {game.Notes}");
        Console.WriteLine($"   Добавлена: {game.AddedDate}");
    }

    public static void Main()
    {
        var collection = new GameCollection();
        try { collection.LoadFromFile("games_data.json"); }
        catch { Console.WriteLine("Не удалось загрузить данные."); }

        while (true)
        {
            Console.WriteLine("\n===== КОЛЛЕКЦИЯ ИГР (C#) =====");
            Console.WriteLine("1. Добавить игру");
            Console.WriteLine("2. Показать все игры");
            Console.WriteLine("3. Показать пройденные игры");
            Console.WriteLine("4. Показать непройденные игры");
            Console.WriteLine("5. Найти игры по названию");
            Console.WriteLine("6. Сортировать по оценке (по убыванию)");
            Console.WriteLine("7. Сортировать по названию");
            Console.WriteLine("8. Редактировать игру");
            Console.WriteLine("9. Удалить игру");
            Console.WriteLine("10. Показать статистику");
            Console.WriteLine("11. Сохранить в файл");
            Console.WriteLine("12. Загрузить из файла");
            Console.WriteLine("13. Экспорт в CSV");
            Console.WriteLine("14. Импорт из CSV");
            Console.WriteLine("0. Выход");
            string choice = ReadString("Выберите действие: ");

            switch (choice)
            {
                case "0": return;
                case "1":
                    string title = ReadString("Название: ");
                    if (string.IsNullOrWhiteSpace(title)) { Console.WriteLine("Название не может быть пустым."); continue; }
                    string genre = ReadString("Жанр: ");
                    if (string.IsNullOrWhiteSpace(genre)) { Console.WriteLine("Жанр не может быть пустым."); continue; }
                    string platform = ReadString("Платформа: ");
                    if (string.IsNullOrWhiteSpace(platform)) { Console.WriteLine("Платформа не может быть пустой."); continue; }
                    int year = ReadInt("Год выпуска: ");
                    bool completed = ReadBool("Статус (1-пройдена, 0-нет): ");
                    int rating = ReadInt("Оценка (1-10): ");
                    string notes = ReadString("Заметки (необязательно): ");
                    try
                    {
                        var game = collection.AddGame(title, genre, platform, year, completed, rating, notes);
                        Console.WriteLine($"Игра добавлена с ID {game.Id}");
                    }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "2":
                    if (!collection.Games.Any()) Console.WriteLine("Нет игр.");
                    else foreach (var g in collection.Games) PrintGame(g);
                    break;
                case "3":
                    var completedGames = collection.FilterByCompleted(true);
                    if (!completedGames.Any()) Console.WriteLine("Нет пройденных игр.");
                    else foreach (var g in completedGames) PrintGame(g);
                    break;
                case "4":
                    var uncompletedGames = collection.FilterByCompleted(false);
                    if (!uncompletedGames.Any()) Console.WriteLine("Нет непройденных игр.");
                    else foreach (var g in uncompletedGames) PrintGame(g);
                    break;
                case "5":
                    string query = ReadString("Введите часть названия: ");
                    var results = collection.SearchGames(query);
                    if (!results.Any()) Console.WriteLine("Игры не найдены.");
                    else foreach (var g in results) PrintGame(g);
                    break;
                case "6":
                    var sortedByRating = collection.SortByRating(false).OrderByDescending(g => g.Rating).ToList();
                    if (!sortedByRating.Any()) Console.WriteLine("Нет игр.");
                    else foreach (var g in sortedByRating) PrintGame(g);
                    break;
                case "7":
                    var sortedByTitle = collection.SortByTitle();
                    if (!sortedByTitle.Any()) Console.WriteLine("Нет игр.");
                    else foreach (var g in sortedByTitle) PrintGame(g);
                    break;
                case "8":
                    int id = ReadInt("Введите ID игры для редактирования: ");
                    var old = collection.FindGame(id);
                    if (old == null) { Console.WriteLine("Игра не найдена."); continue; }
                    Console.WriteLine("Оставьте поле пустым, чтобы не менять.");
                    string newTitle = ReadString($"Название ({old.Title}): ");
                    string newGenre = ReadString($"Жанр ({old.Genre}): ");
                    string newPlatform = ReadString($"Платформа ({old.Platform}): ");
                    string newYearStr = ReadString($"Год ({old.Year}): ");
                    string newCompletedStr = ReadString($"Статус (1-пройдена, 0-нет) сейчас: {(old.Completed ? "1" : "0")}: ");
                    string newRatingStr = ReadString($"Оценка ({old.Rating}): ");
                    string newNotes = ReadString($"Заметки ({old.Notes}): ");
                    var updates = new Dictionary<string, object>();
                    if (!string.IsNullOrWhiteSpace(newTitle)) updates["title"] = newTitle;
                    if (!string.IsNullOrWhiteSpace(newGenre)) updates["genre"] = newGenre;
                    if (!string.IsNullOrWhiteSpace(newPlatform)) updates["platform"] = newPlatform;
                    if (!string.IsNullOrWhiteSpace(newYearStr))
                    {
                        if (int.TryParse(newYearStr, out int y)) updates["year"] = y;
                        else Console.WriteLine("Год должен быть числом, пропускаем.");
                    }
                    if (!string.IsNullOrWhiteSpace(newCompletedStr)) updates["completed"] = newCompletedStr == "1";
                    if (!string.IsNullOrWhiteSpace(newRatingStr))
                    {
                        if (int.TryParse(newRatingStr, out int r)) updates["rating"] = r;
                        else Console.WriteLine("Оценка должна быть числом, пропускаем.");
                    }
                    if (!string.IsNullOrWhiteSpace(newNotes)) updates["notes"] = newNotes;
                    if (collection.EditGame(id, updates)) Console.WriteLine("Игра обновлена.");
                    else Console.WriteLine("Ошибка обновления.");
                    break;
                case "9":
                    int delId = ReadInt("Введите ID игры для удаления: ");
                    if (collection.DeleteGame(delId)) Console.WriteLine("Игра удалена.");
                    else Console.WriteLine("Игра не найдена.");
                    break;
                case "10":
                    var stats = collection.GetStats();
                    Console.WriteLine("\n=== СТАТИСТИКА ===");
                    Console.WriteLine($"Всего игр: {stats["total"]}");
                    Console.WriteLine($"Пройдено: {stats["completed"]}");
                    Console.WriteLine($"Не пройдено: {stats["uncompleted"]}");
                    Console.WriteLine($"Средняя оценка (только пройденные): {stats["avg_rating"]:F2}");
                    Console.WriteLine("По платформам:");
                    var platforms = (Dictionary<string, int>)stats["platforms"];
                    foreach (var kv in platforms) Console.WriteLine($"  {kv.Key}: {kv.Value}");
                    Console.WriteLine("По жанрам:");
                    var genres = (Dictionary<string, int>)stats["genres"];
                    foreach (var kv in genres) Console.WriteLine($"  {kv.Key}: {kv.Value}");
                    break;
                case "11":
                    try { collection.SaveToFile("games_data.json"); Console.WriteLine("Сохранено."); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "12":
                    try { collection.LoadFromFile("games_data.json"); Console.WriteLine("Загружено."); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "13":
                    try { collection.ExportCSV("games_export.csv"); Console.WriteLine("Экспортировано в games_export.csv"); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "14":
                    try { collection.ImportCSV("games_export.csv"); Console.WriteLine("Импортировано из games_export.csv"); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                default: Console.WriteLine("Неизвестная команда."); break;
            }
        }
    }
}
