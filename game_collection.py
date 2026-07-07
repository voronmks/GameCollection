# game_collection.py
import json
import csv
from dataclasses import dataclass, asdict
from datetime import date
from typing import List, Optional
from pathlib import Path

@dataclass
class Game:
    id: int
    title: str
    genre: str
    platform: str
    year: int
    completed: bool
    rating: int  # 1-10
    notes: str
    added_date: str

class GameCollection:
    def __init__(self):
        self.games: List[Game] = []
        self.next_id = 1

    def add_game(self, title: str, genre: str, platform: str, year: int,
                 completed: bool, rating: int, notes: str = "") -> Game:
        if rating < 1 or rating > 10:
            raise ValueError("Оценка должна быть от 1 до 10")
        if year < 1980 or year > date.today().year:
            raise ValueError(f"Год должен быть от 1980 до {date.today().year}")
        if not title or not genre or not platform:
            raise ValueError("Название, жанр и платформа не могут быть пустыми")
        game = Game(
            id=self.next_id,
            title=title,
            genre=genre,
            platform=platform,
            year=year,
            completed=completed,
            rating=rating,
            notes=notes,
            added_date=date.today().isoformat()
        )
        self.games.append(game)
        self.next_id += 1
        return game

    def find_game(self, game_id: int) -> Optional[Game]:
        return next((g for g in self.games if g.id == game_id), None)

    def edit_game(self, game_id: int, **kwargs) -> bool:
        game = self.find_game(game_id)
        if not game:
            return False
        for key, value in kwargs.items():
            if hasattr(game, key) and value is not None:
                setattr(game, key, value)
        return True

    def delete_game(self, game_id: int) -> bool:
        game = self.find_game(game_id)
        if game:
            self.games.remove(game)
            return True
        return False

    def search_games(self, query: str) -> List[Game]:
        q = query.lower()
        return [g for g in self.games if q in g.title.lower()]

    def filter_by_completed(self, completed: bool) -> List[Game]:
        return [g for g in self.games if g.completed == completed]

    def filter_by_genre(self, genre: str) -> List[Game]:
        return [g for g in self.games if g.genre.lower() == genre.lower()]

    def filter_by_platform(self, platform: str) -> List[Game]:
        return [g for g in self.games if g.platform.lower() == platform.lower()]

    def sort_by_rating(self, reverse: bool = True) -> List[Game]:
        return sorted(self.games, key=lambda g: g.rating, reverse=reverse)

    def sort_by_title(self) -> List[Game]:
        return sorted(self.games, key=lambda g: g.title.lower())

    def get_stats(self) -> dict:
        total = len(self.games)
        completed_count = len(self.filter_by_completed(True))
        uncompleted = total - completed_count
        ratings = [g.rating for g in self.games if g.completed]
        avg_rating = sum(ratings) / len(ratings) if ratings else 0.0
        platforms = {}
        genres = {}
        for g in self.games:
            platforms[g.platform] = platforms.get(g.platform, 0) + 1
            genres[g.genre] = genres.get(g.genre, 0) + 1
        return {
            "total": total,
            "completed": completed_count,
            "uncompleted": uncompleted,
            "avg_rating": avg_rating,
            "platforms": platforms,
            "genres": genres
        }

    def save_to_file(self, filename: str = "games_data.json") -> None:
        data = {"games": [asdict(g) for g in self.games]}
        with open(filename, "w", encoding="utf-8") as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def load_from_file(self, filename: str = "games_data.json") -> None:
        path = Path(filename)
        if not path.exists():
            return
        with open(filename, "r", encoding="utf-8") as f:
            data = json.load(f)
            self.games.clear()
            for item in data.get("games", []):
                game = Game(
                    id=item["id"],
                    title=item["title"],
                    genre=item["genre"],
                    platform=item["platform"],
                    year=item["year"],
                    completed=item["completed"],
                    rating=item["rating"],
                    notes=item.get("notes", ""),
                    added_date=item["added_date"]
                )
                self.games.append(game)
                if game.id >= self.next_id:
                    self.next_id = game.id + 1

    def export_csv(self, filename: str = "games_export.csv") -> None:
        with open(filename, "w", newline="", encoding="utf-8") as f:
            writer = csv.writer(f, delimiter=";")
            writer.writerow(["ID", "Название", "Жанр", "Платформа", "Год", "Пройдена", "Оценка", "Заметки", "Дата добавления"])
            for g in self.games:
                writer.writerow([g.id, g.title, g.genre, g.platform, g.year,
                                 "Да" if g.completed else "Нет", g.rating, g.notes, g.added_date])

    def import_csv(self, filename: str = "games_export.csv") -> None:
        path = Path(filename)
        if not path.exists():
            raise FileNotFoundError("Файл не найден")
        with open(filename, "r", encoding="utf-8") as f:
            reader = csv.DictReader(f, delimiter=";")
            for row in reader:
                try:
                    self.add_game(
                        title=row["Название"],
                        genre=row["Жанр"],
                        platform=row["Платформа"],
                        year=int(row["Год"]),
                        completed=row["Пройдена"] == "Да",
                        rating=int(row["Оценка"]),
                        notes=row["Заметки"]
                    )
                except Exception as e:
                    print(f"Ошибка импорта строки: {e}")

def print_game(game: Game) -> None:
    status = "✅ Пройдена" if game.completed else "⏳ Не пройдена"
    print(f"#{game.id} - {game.title} ({game.year})")
    print(f"   Жанр: {game.genre}, Платформа: {game.platform}")
    print(f"   {status}, Оценка: {game.rating}/10")
    if game.notes:
        print(f"   Заметки: {game.notes}")
    print(f"   Добавлена: {game.added_date}")

def main():
    collection = GameCollection()
    collection.load_from_file()

    while True:
        print("\n===== КОЛЛЕКЦИЯ ИГР (Python) =====")
        print("1. Добавить игру")
        print("2. Показать все игры")
        print("3. Показать пройденные игры")
        print("4. Показать непройденные игры")
        print("5. Найти игры по названию")
        print("6. Сортировать по оценке (по убыванию)")
        print("7. Сортировать по названию")
        print("8. Редактировать игру")
        print("9. Удалить игру")
        print("10. Показать статистику")
        print("11. Сохранить в файл")
        print("12. Загрузить из файла")
        print("13. Экспорт в CSV")
        print("14. Импорт из CSV")
        print("0. Выход")
        choice = input("Выберите действие: ").strip()

        if choice == "0":
            break
        elif choice == "1":
            title = input("Название: ").strip()
            if not title:
                print("Название не может быть пустым.")
                continue
            genre = input("Жанр: ").strip()
            if not genre:
                print("Жанр не может быть пустым.")
                continue
            platform = input("Платформа: ").strip()
            if not platform:
                print("Платформа не может быть пустой.")
                continue
            try:
                year = int(input("Год выпуска: ").strip())
            except ValueError:
                print("Введите число.")
                continue
            completed_input = input("Статус (1-пройдена, 0-нет): ").strip()
            completed = completed_input == "1"
            try:
                rating = int(input("Оценка (1-10): ").strip())
            except ValueError:
                rating = 0
            notes = input("Заметки (необязательно): ").strip()
            try:
                game = collection.add_game(title, genre, platform, year, completed, rating, notes)
                print(f"Игра добавлена с ID {game.id}")
            except Exception as e:
                print("Ошибка:", e)
        elif choice == "2":
            if not collection.games:
                print("Нет игр.")
            else:
                for g in collection.games:
                    print_game(g)
        elif choice == "3":
            games = collection.filter_by_completed(True)
            if not games:
                print("Нет пройденных игр.")
            else:
                for g in games:
                    print_game(g)
        elif choice == "4":
            games = collection.filter_by_completed(False)
            if not games:
                print("Нет непройденных игр.")
            else:
                for g in games:
                    print_game(g)
        elif choice == "5":
            query = input("Введите часть названия: ").strip()
            if not query:
                print("Введите текст.")
                continue
            results = collection.search_games(query)
            if not results:
                print("Игры не найдены.")
            else:
                for g in results:
                    print_game(g)
        elif choice == "6":
            sorted_games = collection.sort_by_rating(reverse=True)
            if not sorted_games:
                print("Нет игр.")
            else:
                for g in sorted_games:
                    print_game(g)
        elif choice == "7":
            sorted_games = collection.sort_by_title()
            if not sorted_games:
                print("Нет игр.")
            else:
                for g in sorted_games:
                    print_game(g)
        elif choice == "8":
            try:
                gid = int(input("Введите ID игры для редактирования: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            game = collection.find_game(gid)
            if not game:
                print("Игра не найдена.")
                continue
            print("Оставьте поле пустым, чтобы не менять.")
            new_title = input(f"Название ({game.title}): ").strip()
            new_genre = input(f"Жанр ({game.genre}): ").strip()
            new_platform = input(f"Платформа ({game.platform}): ").strip()
            new_year = input(f"Год ({game.year}): ").strip()
            new_completed = input(f"Статус (1-пройдена, 0-нет) сейчас: {'1' if game.completed else '0'}: ").strip()
            new_rating = input(f"Оценка ({game.rating}): ").strip()
            new_notes = input(f"Заметки ({game.notes}): ").strip()
            updates = {}
            if new_title: updates["title"] = new_title
            if new_genre: updates["genre"] = new_genre
            if new_platform: updates["platform"] = new_platform
            if new_year:
                try:
                    updates["year"] = int(new_year)
                except ValueError:
                    print("Год должен быть числом, пропускаем.")
            if new_completed: updates["completed"] = new_completed == "1"
            if new_rating:
                try:
                    updates["rating"] = int(new_rating)
                except ValueError:
                    print("Оценка должна быть числом, пропускаем.")
            if new_notes: updates["notes"] = new_notes
            if collection.edit_game(gid, **updates):
                print("Игра обновлена.")
            else:
                print("Ошибка обновления.")
        elif choice == "9":
            try:
                gid = int(input("Введите ID игры для удаления: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            if collection.delete_game(gid):
                print("Игра удалена.")
            else:
                print("Игра не найдена.")
        elif choice == "10":
            stats = collection.get_stats()
            print("\n=== СТАТИСТИКА ===")
            print(f"Всего игр: {stats['total']}")
            print(f"Пройдено: {stats['completed']}")
            print(f"Не пройдено: {stats['uncompleted']}")
            print(f"Средняя оценка (только пройденные): {stats['avg_rating']:.2f}")
            print("По платформам:")
            for p, c in stats['platforms'].items():
                print(f"  {p}: {c}")
            print("По жанрам:")
            for g, c in stats['genres'].items():
                print(f"  {g}: {c}")
        elif choice == "11":
            collection.save_to_file()
            print("Сохранено.")
        elif choice == "12":
            collection.load_from_file()
            print("Загружено.")
        elif choice == "13":
            collection.export_csv()
            print("Экспортировано в games_export.csv")
        elif choice == "14":
            try:
                collection.import_csv()
                print("Импортировано из games_export.csv")
            except FileNotFoundError:
                print("Файл games_export.csv не найден.")
            except Exception as e:
                print("Ошибка импорта:", e)
        else:
            print("Неизвестная команда.")

if __name__ == "__main__":
    main()
