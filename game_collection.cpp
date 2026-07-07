// game_collection.cpp
#include <iostream>
#include <vector>
#include <string>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <iomanip>
#include <ctime>
#include <map>
#include <variant>
#include <regex>
#include <cctype>

using namespace std;

struct Game {
    int id;
    string title;
    string genre;
    string platform;
    int year;
    bool completed;
    int rating;
    string notes;
    string addedDate;

    Game(int id, const string& title, const string& genre, const string& platform,
         int year, bool completed, int rating, const string& notes = "", const string& addedDate = "")
        : id(id), title(title), genre(genre), platform(platform), year(year),
          completed(completed), rating(rating), notes(notes), addedDate(addedDate) {
        if (addedDate.empty()) {
            time_t now = time(nullptr);
            tm* tm_now = localtime(&now);
            char buf[11];
            strftime(buf, sizeof(buf), "%Y-%m-%d", tm_now);
            this->addedDate = string(buf);
        }
    }
};

class GameCollection {
private:
    vector<Game> games;
    int nextId = 1;

    int currentYear() {
        time_t now = time(nullptr);
        tm* tm_now = localtime(&now);
        return tm_now->tm_year + 1900;
    }

public:
    Game addGame(const string& title, const string& genre, const string& platform,
                 int year, bool completed, int rating, const string& notes = "") {
        if (rating < 1 || rating > 10) throw invalid_argument("Оценка должна быть от 1 до 10");
        int cy = currentYear();
        if (year < 1980 || year > cy) throw invalid_argument("Год должен быть от 1980 до " + to_string(cy));
        if (title.empty() || genre.empty() || platform.empty())
            throw invalid_argument("Название, жанр и платформа не могут быть пустыми");
        Game game(nextId, title, genre, platform, year, completed, rating, notes);
        games.push_back(game);
        nextId++;
        return game;
    }

    Game* findGame(int id) {
        auto it = find_if(games.begin(), games.end(), [id](const Game& g) { return g.id == id; });
        return it != games.end() ? &(*it) : nullptr;
    }

    bool editGame(int id, const map<string, string>& updates) {
        Game* game = findGame(id);
        if (!game) return false;
        for (const auto& [key, value] : updates) {
            if (key == "title") game->title = value;
            else if (key == "genre") game->genre = value;
            else if (key == "platform") game->platform = value;
            else if (key == "year") game->year = stoi(value);
            else if (key == "completed") game->completed = (value == "1");
            else if (key == "rating") game->rating = stoi(value);
            else if (key == "notes") game->notes = value;
        }
        return true;
    }

    bool deleteGame(int id) {
        auto it = find_if(games.begin(), games.end(), [id](const Game& g) { return g.id == id; });
        if (it == games.end()) return false;
        games.erase(it);
        return true;
    }

    vector<Game> searchGames(const string& query) {
        string q = query;
        transform(q.begin(), q.end(), q.begin(), ::tolower);
        vector<Game> result;
        for (const auto& g : games) {
            string titleLower = g.title;
            transform(titleLower.begin(), titleLower.end(), titleLower.begin(), ::tolower);
            if (titleLower.find(q) != string::npos) result.push_back(g);
        }
        return result;
    }

    vector<Game> filterByCompleted(bool completed) const {
        vector<Game> result;
        for (const auto& g : games) {
            if (g.completed == completed) result.push_back(g);
        }
        return result;
    }

    vector<Game> filterByGenre(const string& genre) const {
        vector<Game> result;
        for (const auto& g : games) {
            if (g.genre == genre) result.push_back(g);
        }
        return result;
    }

    vector<Game> filterByPlatform(const string& platform) const {
        vector<Game> result;
        for (const auto& g : games) {
            if (g.platform == platform) result.push_back(g);
        }
        return result;
    }

    vector<Game> sortByRating(bool reverse = true) {
        vector<Game> sorted = games;
        sort(sorted.begin(), sorted.end(), [reverse](const Game& a, const Game& b) {
            return reverse ? a.rating > b.rating : a.rating < b.rating;
        });
        return sorted;
    }

    vector<Game> sortByTitle() {
        vector<Game> sorted = games;
        sort(sorted.begin(), sorted.end(), [](const Game& a, const Game& b) {
            string al = a.title, bl = b.title;
            transform(al.begin(), al.end(), al.begin(), ::tolower);
            transform(bl.begin(), bl.end(), bl.begin(), ::tolower);
            return al < bl;
        });
        return sorted;
    }

    map<string, variant<int, double, map<string, int>>> getStats() const {
        int total = games.size();
        int completedCount = filterByCompleted(true).size();
        int uncompleted = total - completedCount;
        int sumRating = 0;
        int ratingCount = 0;
        for (const auto& g : games) {
            if (g.completed) { sumRating += g.rating; ratingCount++; }
        }
        double avgRating = ratingCount > 0 ? static_cast<double>(sumRating) / ratingCount : 0.0;
        map<string, int> platforms, genres;
        for (const auto& g : games) {
            platforms[g.platform]++;
            genres[g.genre]++;
        }
        map<string, variant<int, double, map<string, int>>> stats;
        stats["total"] = total;
        stats["completed"] = completedCount;
        stats["uncompleted"] = uncompleted;
        stats["avg_rating"] = avgRating;
        stats["platforms"] = platforms;
        stats["genres"] = genres;
        return stats;
    }

    void saveToFile(const string& filename = "games_data.txt") {
        ofstream out(filename);
        if (!out) return;
        for (const auto& g : games) {
            out << g.id << '|'
                << g.title << '|'
                << g.genre << '|'
                << g.platform << '|'
                << g.year << '|'
                << g.completed << '|'
                << g.rating << '|'
                << g.notes << '|'
                << g.addedDate << '\n';
        }
    }

    void loadFromFile(const string& filename = "games_data.txt") {
        ifstream in(filename);
        if (!in) return;
        games.clear();
        string line;
        while (getline(in, line)) {
            stringstream ss(line);
            string idStr, title, genre, platform, yearStr, completedStr, ratingStr, notes, addedDate;
            getline(ss, idStr, '|');
            getline(ss, title, '|');
            getline(ss, genre, '|');
            getline(ss, platform, '|');
            getline(ss, yearStr, '|');
            getline(ss, completedStr, '|');
            getline(ss, ratingStr, '|');
            getline(ss, notes, '|');
            getline(ss, addedDate, '|');
            int id = stoi(idStr);
            int year = stoi(yearStr);
            bool completed = (completedStr == "1");
            int rating = stoi(ratingStr);
            games.emplace_back(id, title, genre, platform, year, completed, rating, notes, addedDate);
            if (id >= nextId) nextId = id + 1;
        }
    }

    void exportCSV(const string& filename = "games_export.csv") {
        ofstream out(filename);
        if (!out) return;
        out << "ID;Название;Жанр;Платформа;Год;Пройдена;Оценка;Заметки;Дата добавления\n";
        for (const auto& g : games) {
            out << g.id << ';'
                << g.title << ';'
                << g.genre << ';'
                << g.platform << ';'
                << g.year << ';'
                << (g.completed ? "Да" : "Нет") << ';'
                << g.rating << ';'
                << g.notes << ';'
                << g.addedDate << '\n';
        }
    }

    void importCSV(const string& filename = "games_export.csv") {
        ifstream in(filename);
        if (!in) return;
        string line;
        getline(in, line); // skip header
        while (getline(in, line)) {
            stringstream ss(line);
            string idStr, title, genre, platform, yearStr, completedStr, ratingStr, notes, addedDate;
            getline(ss, idStr, ';');
            getline(ss, title, ';');
            getline(ss, genre, ';');
            getline(ss, platform, ';');
            getline(ss, yearStr, ';');
            getline(ss, completedStr, ';');
            getline(ss, ratingStr, ';');
            getline(ss, notes, ';');
            getline(ss, addedDate, ';');
            try {
                addGame(title, genre, platform, stoi(yearStr),
                        completedStr == "Да", stoi(ratingStr), notes);
            } catch (const exception& e) {
                cout << "Ошибка импорта строки: " << e.what() << "\n";
            }
        }
    }

    const vector<Game>& getGames() const { return games; }
};

string readString(const string& prompt) {
    cout << prompt;
    string input;
    getline(cin, input);
    return input;
}

int readInt(const string& prompt) {
    while (true) {
        cout << prompt;
        string input;
        getline(cin, input);
        try {
            return stoi(input);
        } catch (...) {
            cout << "Введите число.\n";
        }
    }
}

bool readBool(const string& prompt) {
    while (true) {
        string input = readString(prompt);
        if (input == "1") return true;
        if (input == "0") return false;
        cout << "Введите 1 или 0.\n";
    }
}

void printGame(const Game& game) {
    string status = game.completed ? "✅ Пройдена" : "⏳ Не пройдена";
    cout << "#" << game.id << " - " << game.title << " (" << game.year << ")\n";
    cout << "   Жанр: " << game.genre << ", Платформа: " << game.platform << "\n";
    cout << "   " << status << ", Оценка: " << game.rating << "/10\n";
    if (!game.notes.empty()) cout << "   Заметки: " << game.notes << "\n";
    cout << "   Добавлена: " << game.addedDate << "\n";
}

int main() {
    GameCollection collection;
    collection.loadFromFile();

    while (true) {
        cout << "\n===== КОЛЛЕКЦИЯ ИГР (C++) =====" << endl;
        cout << "1. Добавить игру\n";
        cout << "2. Показать все игры\n";
        cout << "3. Показать пройденные игры\n";
        cout << "4. Показать непройденные игры\n";
        cout << "5. Найти игры по названию\n";
        cout << "6. Сортировать по оценке (по убыванию)\n";
        cout << "7. Сортировать по названию\n";
        cout << "8. Редактировать игру\n";
        cout << "9. Удалить игру\n";
        cout << "10. Показать статистику\n";
        cout << "11. Сохранить в файл\n";
        cout << "12. Загрузить из файла\n";
        cout << "13. Экспорт в CSV\n";
        cout << "14. Импорт из CSV\n";
        cout << "0. Выход\n";
        string choice = readString("Выберите действие: ");

        if (choice == "0") break;

        if (choice == "1") {
            string title = readString("Название: ");
            if (title.empty()) { cout << "Название не может быть пустым.\n"; continue; }
            string genre = readString("Жанр: ");
            if (genre.empty()) { cout << "Жанр не может быть пустым.\n"; continue; }
            string platform = readString("Платформа: ");
            if (platform.empty()) { cout << "Платформа не может быть пустой.\n"; continue; }
            int year = readInt("Год выпуска: ");
            bool completed = readBool("Статус (1-пройдена, 0-нет): ");
            int rating = readInt("Оценка (1-10): ");
            string notes = readString("Заметки (необязательно): ");
            try {
                Game game = collection.addGame(title, genre, platform, year, completed, rating, notes);
                cout << "Игра добавлена с ID " << game.id << "\n";
            } catch (const exception& e) {
                cout << "Ошибка: " << e.what() << "\n";
            }
        } else if (choice == "2") {
            if (collection.getGames().empty()) {
                cout << "Нет игр.\n";
            } else {
                for (const auto& g : collection.getGames()) printGame(g);
            }
        } else if (choice == "3") {
            auto games = collection.filterByCompleted(true);
            if (games.empty()) cout << "Нет пройденных игр.\n";
            else for (const auto& g : games) printGame(g);
        } else if (choice == "4") {
            auto games = collection.filterByCompleted(false);
            if (games.empty()) cout << "Нет непройденных игр.\n";
            else for (const auto& g : games) printGame(g);
        } else if (choice == "5") {
            string query = readString("Введите часть названия: ");
            auto results = collection.searchGames(query);
            if (results.empty()) cout << "Игры не найдены.\n";
            else for (const auto& g : results) printGame(g);
        } else if (choice == "6") {
            auto sorted = collection.sortByRating(true);
            if (sorted.empty()) cout << "Нет игр.\n";
            else for (const auto& g : sorted) printGame(g);
        } else if (choice == "7") {
            auto sorted = collection.sortByTitle();
            if (sorted.empty()) cout << "Нет игр.\n";
            else for (const auto& g : sorted) printGame(g);
        } else if (choice == "8") {
            int id = readInt("Введите ID игры для редактирования: ");
            Game* game = collection.findGame(id);
            if (!game) { cout << "Игра не найдена.\n"; continue; }
            cout << "Оставьте поле пустым, чтобы не менять.\n";
            string newTitle = readString("Название (" + game->title + "): ");
            string newGenre = readString("Жанр (" + game->genre + "): ");
            string newPlatform = readString("Платформа (" + game->platform + "): ");
            string newYear = readString("Год (" + to_string(game->year) + "): ");
            string newCompleted = readString("Статус (1-пройдена, 0-нет) сейчас: " + string(game->completed ? "1" : "0") + ": ");
            string newRating = readString("Оценка (" + to_string(game->rating) + "): ");
            string newNotes = readString("Заметки (" + game->notes + "): ");
            map<string, string> updates;
            if (!newTitle.empty()) updates["title"] = newTitle;
            if (!newGenre.empty()) updates["genre"] = newGenre;
            if (!newPlatform.empty()) updates["platform"] = newPlatform;
            if (!newYear.empty()) updates["year"] = newYear;
            if (!newCompleted.empty()) updates["completed"] = newCompleted;
            if (!newRating.empty()) updates["rating"] = newRating;
            if (!newNotes.empty()) updates["notes"] = newNotes;
            if (collection.editGame(id, updates)) {
                cout << "Игра обновлена.\n";
            } else {
                cout << "Ошибка обновления.\n";
            }
        } else if (choice == "9") {
            int id = readInt("Введите ID игры для удаления: ");
            if (collection.deleteGame(id)) {
                cout << "Игра удалена.\n";
            } else {
                cout << "Игра не найдена.\n";
            }
        } else if (choice == "10") {
            auto stats = collection.getStats();
            cout << "\n=== СТАТИСТИКА ===\n";
            cout << "Всего игр: " << get<int>(stats["total"]) << "\n";
            cout << "Пройдено: " << get<int>(stats["completed"]) << "\n";
            cout << "Не пройдено: " << get<int>(stats["uncompleted"]) << "\n";
            cout << "Средняя оценка (только пройденные): " << fixed << setprecision(2) << get<double>(stats["avg_rating"]) << "\n";
            cout << "По платформам:\n";
            auto platforms = get<map<string, int>>(stats["platforms"]);
            for (const auto& [p, c] : platforms) cout << "  " << p << ": " << c << "\n";
            cout << "По жанрам:\n";
            auto genres = get<map<string, int>>(stats["genres"]);
            for (const auto& [g, c] : genres) cout << "  " << g << ": " << c << "\n";
        } else if (choice == "11") {
            collection.saveToFile();
            cout << "Сохранено.\n";
        } else if (choice == "12") {
            collection.loadFromFile();
            cout << "Загружено.\n";
        } else if (choice == "13") {
            collection.exportCSV();
            cout << "Экспортировано в games_export.csv\n";
        } else if (choice == "14") {
            try {
                collection.importCSV();
                cout << "Импортировано из games_export.csv\n";
            } catch (const exception& e) {
                cout << "Ошибка импорта: " << e.what() << "\n";
            }
        } else {
            cout << "Неизвестная команда.\n";
        }
    }
    return 0;
}
