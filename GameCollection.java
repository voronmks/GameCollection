// GameCollection.java
import java.io.*;
import java.nio.file.*;
import java.time.LocalDate;
import java.util.*;
import java.util.stream.Collectors;

record Game(int id, String title, String genre, String platform, int year, boolean completed, int rating, String notes, String addedDate) implements Serializable {}

class GamesData implements Serializable {
    private static final long serialVersionUID = 1L;
    public List<Game> games;
}

class GameCollection implements Serializable {
    private static final long serialVersionUID = 1L;
    private List<Game> games = new ArrayList<>();
    private int nextId = 1;

    public Game addGame(String title, String genre, String platform, int year, boolean completed, int rating, String notes) {
        if (rating < 1 || rating > 10) throw new IllegalArgumentException("Оценка должна быть от 1 до 10");
        if (year < 1980 || year > LocalDate.now().getYear())
            throw new IllegalArgumentException("Год должен быть от 1980 до " + LocalDate.now().getYear());
        if (title.isBlank() || genre.isBlank() || platform.isBlank())
            throw new IllegalArgumentException("Название, жанр и платформа не могут быть пустыми");
        if (notes == null) notes = "";
        Game game = new Game(nextId, title, genre, platform, year, completed, rating, notes, LocalDate.now().toString());
        games.add(game);
        nextId++;
        return game;
    }

    public Optional<Game> findGame(int id) {
        return games.stream().filter(g -> g.id() == id).findFirst();
    }

    public boolean editGame(int id, Map<String, Object> updates) {
        Optional<Game> opt = findGame(id);
        if (opt.isEmpty()) return false;
        Game old = opt.get();
        games.remove(old);
        String title = (String) updates.getOrDefault("title", old.title());
        String genre = (String) updates.getOrDefault("genre", old.genre());
        String platform = (String) updates.getOrDefault("platform", old.platform());
        int year = (int) updates.getOrDefault("year", old.year());
        boolean completed = (boolean) updates.getOrDefault("completed", old.completed());
        int rating = (int) updates.getOrDefault("rating", old.rating());
        String notes = (String) updates.getOrDefault("notes", old.notes());
        Game updated = new Game(old.id(), title, genre, platform, year, completed, rating, notes, old.addedDate());
        games.add(updated);
        return true;
    }

    public boolean deleteGame(int id) {
        return games.removeIf(g -> g.id() == id);
    }

    public List<Game> searchGames(String query) {
        String q = query.toLowerCase();
        return games.stream()
                .filter(g -> g.title().toLowerCase().contains(q))
                .collect(Collectors.toList());
    }

    public List<Game> filterByCompleted(boolean completed) {
        return games.stream().filter(g -> g.completed() == completed).collect(Collectors.toList());
    }

    public List<Game> filterByGenre(String genre) {
        return games.stream().filter(g -> g.genre().equalsIgnoreCase(genre)).collect(Collectors.toList());
    }

    public List<Game> filterByPlatform(String platform) {
        return games.stream().filter(g -> g.platform().equalsIgnoreCase(platform)).collect(Collectors.toList());
    }

    public List<Game> sortByRating(boolean reverse) {
        return games.stream()
                .sorted((a, b) -> reverse ? Integer.compare(b.rating(), a.rating()) : Integer.compare(a.rating(), b.rating()))
                .collect(Collectors.toList());
    }

    public List<Game> sortByTitle() {
        return games.stream()
                .sorted(Comparator.comparing(g -> g.title().toLowerCase()))
                .collect(Collectors.toList());
    }

    public Map<String, Object> getStats() {
        int total = games.size();
        int completedCount = filterByCompleted(true).size();
        int uncompleted = total - completedCount;
        double avgRating = filterByCompleted(true).stream().mapToInt(Game::rating).average().orElse(0);
        Map<String, Integer> platforms = new HashMap<>();
        Map<String, Integer> genres = new HashMap<>();
        games.forEach(g -> {
            platforms.put(g.platform(), platforms.getOrDefault(g.platform(), 0) + 1);
            genres.put(g.genre(), genres.getOrDefault(g.genre(), 0) + 1);
        });
        Map<String, Object> stats = new HashMap<>();
        stats.put("total", total);
        stats.put("completed", completedCount);
        stats.put("uncompleted", uncompleted);
        stats.put("avg_rating", avgRating);
        stats.put("platforms", platforms);
        stats.put("genres", genres);
        return stats;
    }

    public void saveToFile(String filename) throws IOException {
        GamesData data = new GamesData();
        data.games = new ArrayList<>(games);
        try (ObjectOutputStream oos = new ObjectOutputStream(Files.newOutputStream(Paths.get(filename)))) {
            oos.writeObject(data);
        }
    }

    public void loadFromFile(String filename) throws IOException, ClassNotFoundException {
        try (ObjectInputStream ois = new ObjectInputStream(Files.newInputStream(Paths.get(filename)))) {
            GamesData data = (GamesData) ois.readObject();
            games = new ArrayList<>(data.games);
            for (Game g : games) {
                if (g.id() >= nextId) nextId = g.id() + 1;
            }
        }
    }

    public void exportCSV(String filename) throws IOException {
        try (PrintWriter pw = new PrintWriter(Files.newBufferedWriter(Paths.get(filename)))) {
            pw.println("ID;Название;Жанр;Платформа;Год;Пройдена;Оценка;Заметки;Дата добавления");
            for (Game g : games) {
                pw.printf("%d;%s;%s;%s;%d;%s;%d;%s;%s%n",
                        g.id(), g.title(), g.genre(), g.platform(), g.year(),
                        g.completed() ? "Да" : "Нет", g.rating(), g.notes(), g.addedDate());
            }
        }
    }

    public void importCSV(String filename) throws IOException {
        try (BufferedReader br = Files.newBufferedReader(Paths.get(filename))) {
            String line = br.readLine(); // skip header
            while ((line = br.readLine()) != null) {
                String[] parts = line.split(";");
                if (parts.length < 9) continue;
                String title = parts[1];
                String genre = parts[2];
                String platform = parts[3];
                int year = Integer.parseInt(parts[4]);
                boolean completed = parts[5].equals("Да");
                int rating = Integer.parseInt(parts[6]);
                String notes = parts[7];
                try {
                    addGame(title, genre, platform, year, completed, rating, notes);
                } catch (Exception e) {
                    System.out.println("Ошибка импорта строки: " + e.getMessage());
                }
            }
        }
    }

    public List<Game> getGames() { return Collections.unmodifiableList(games); }
}

public class GameCollectionApp {
    private static final Scanner scanner = new Scanner(System.in);

    private static String readString(String prompt) {
        System.out.print(prompt);
        return scanner.nextLine().trim();
    }

    private static int readInt(String prompt) {
        while (true) {
            try {
                System.out.print(prompt);
                return Integer.parseInt(scanner.nextLine().trim());
            } catch (NumberFormatException e) {
                System.out.println("Введите число.");
            }
        }
    }

    private static boolean readBool(String prompt) {
        while (true) {
            String input = readString(prompt);
            if (input.equals("1")) return true;
            if (input.equals("0")) return false;
            System.out.println("Введите 1 или 0.");
        }
    }

    private static void printGame(Game game) {
        String status = game.completed() ? "✅ Пройдена" : "⏳ Не пройдена";
        System.out.printf("#%d - %s (%d)%n", game.id(), game.title(), game.year());
        System.out.printf("   Жанр: %s, Платформа: %s%n", game.genre(), game.platform());
        System.out.printf("   %s, Оценка: %d/10%n", status, game.rating());
        if (!game.notes().isBlank()) {
            System.out.printf("   Заметки: %s%n", game.notes());
        }
        System.out.printf("   Добавлена: %s%n", game.addedDate());
    }

    public static void main(String[] args) {
        GameCollection collection = new GameCollection();
        try {
            collection.loadFromFile("games_data.ser");
        } catch (IOException | ClassNotFoundException e) {
            System.out.println("Не удалось загрузить данные.");
        }

        while (true) {
            System.out.println("\n===== КОЛЛЕКЦИЯ ИГР (Java) =====");
            System.out.println("1. Добавить игру");
            System.out.println("2. Показать все игры");
            System.out.println("3. Показать пройденные игры");
            System.out.println("4. Показать непройденные игры");
            System.out.println("5. Найти игры по названию");
            System.out.println("6. Сортировать по оценке (по убыванию)");
            System.out.println("7. Сортировать по названию");
            System.out.println("8. Редактировать игру");
            System.out.println("9. Удалить игру");
            System.out.println("10. Показать статистику");
            System.out.println("11. Сохранить в файл");
            System.out.println("12. Загрузить из файла");
            System.out.println("13. Экспорт в CSV");
            System.out.println("14. Импорт из CSV");
            System.out.println("0. Выход");
            String choice = readString("Выберите действие: ");

            switch (choice) {
                case "0" -> { return; }
                case "1" -> {
                    String title = readString("Название: ");
                    if (title.isBlank()) { System.out.println("Название не может быть пустым."); continue; }
                    String genre = readString("Жанр: ");
                    if (genre.isBlank()) { System.out.println("Жанр не может быть пустым."); continue; }
                    String platform = readString("Платформа: ");
                    if (platform.isBlank()) { System.out.println("Платформа не может быть пустой."); continue; }
                    int year = readInt("Год выпуска: ");
                    boolean completed = readBool("Статус (1-пройдена, 0-нет): ");
                    int rating = readInt("Оценка (1-10): ");
                    String notes = readString("Заметки (необязательно): ");
                    try {
                        Game game = collection.addGame(title, genre, platform, year, completed, rating, notes);
                        System.out.println("Игра добавлена с ID " + game.id());
                    } catch (Exception e) {
                        System.out.println("Ошибка: " + e.getMessage());
                    }
                }
                case "2" -> {
                    if (collection.getGames().isEmpty()) System.out.println("Нет игр.");
                    else collection.getGames().forEach(GameCollectionApp::printGame);
                }
                case "3" -> {
                    var games = collection.filterByCompleted(true);
                    if (games.isEmpty()) System.out.println("Нет пройденных игр.");
                    else games.forEach(GameCollectionApp::printGame);
                }
                case "4" -> {
                    var games = collection.filterByCompleted(false);
                    if (games.isEmpty()) System.out.println("Нет непройденных игр.");
                    else games.forEach(GameCollectionApp::printGame);
                }
                case "5" -> {
                    String query = readString("Введите часть названия: ");
                    var results = collection.searchGames(query);
                    if (results.isEmpty()) System.out.println("Игры не найдены.");
                    else results.forEach(GameCollectionApp::printGame);
                }
                case "6" -> {
                    var sorted = collection.sortByRating(true);
                    if (sorted.isEmpty()) System.out.println("Нет игр.");
                    else sorted.forEach(GameCollectionApp::printGame);
                }
                case "7" -> {
                    var sorted = collection.sortByTitle();
                    if (sorted.isEmpty()) System.out.println("Нет игр.");
                    else sorted.forEach(GameCollectionApp::printGame);
                }
                case "8" -> {
                    int id = readInt("Введите ID игры для редактирования: ");
                    var opt = collection.findGame(id);
                    if (opt.isEmpty()) { System.out.println("Игра не найдена."); continue; }
                    Game old = opt.get();
                    System.out.println("Оставьте поле пустым, чтобы не менять.");
                    String newTitle = readString("Название (" + old.title() + "): ");
                    String newGenre = readString("Жанр (" + old.genre() + "): ");
                    String newPlatform = readString("Платформа (" + old.platform() + "): ");
                    String newYearStr = readString("Год (" + old.year() + "): ");
                    String newCompletedStr = readString("Статус (1-пройдена, 0-нет) сейчас: " + (old.completed() ? "1" : "0") + ": ");
                    String newRatingStr = readString("Оценка (" + old.rating() + "): ");
                    String newNotes = readString("Заметки (" + old.notes() + "): ");
                    Map<String, Object> updates = new HashMap<>();
                    if (!newTitle.isBlank()) updates.put("title", newTitle);
                    if (!newGenre.isBlank()) updates.put("genre", newGenre);
                    if (!newPlatform.isBlank()) updates.put("platform", newPlatform);
                    if (!newYearStr.isBlank()) {
                        try { updates.put("year", Integer.parseInt(newYearStr)); }
                        catch (NumberFormatException e) { System.out.println("Год должен быть числом, пропускаем."); }
                    }
                    if (!newCompletedStr.isBlank()) updates.put("completed", newCompletedStr.equals("1"));
                    if (!newRatingStr.isBlank()) {
                        try { updates.put("rating", Integer.parseInt(newRatingStr)); }
                        catch (NumberFormatException e) { System.out.println("Оценка должна быть числом, пропускаем."); }
                    }
                    if (!newNotes.isBlank()) updates.put("notes", newNotes);
                    if (collection.editGame(id, updates)) System.out.println("Игра обновлена.");
                    else System.out.println("Ошибка обновления.");
                }
                case "9" -> {
                    int id = readInt("Введите ID игры для удаления: ");
                    if (collection.deleteGame(id)) System.out.println("Игра удалена.");
                    else System.out.println("Игра не найдена.");
                }
                case "10" -> {
                    var stats = collection.getStats();
                    System.out.println("\n=== СТАТИСТИКА ===");
                    System.out.println("Всего игр: " + stats.get("total"));
                    System.out.println("Пройдено: " + stats.get("completed"));
                    System.out.println("Не пройдено: " + stats.get("uncompleted"));
                    System.out.printf("Средняя оценка (только пройденные): %.2f%n", stats.get("avg_rating"));
                    System.out.println("По платформам:");
                    @SuppressWarnings("unchecked")
                    Map<String, Integer> platforms = (Map<String, Integer>) stats.get("platforms");
                    platforms.forEach((p, c) -> System.out.println("  " + p + ": " + c));
                    System.out.println("По жанрам:");
                    @SuppressWarnings("unchecked")
                    Map<String, Integer> genres = (Map<String, Integer>) stats.get("genres");
                    genres.forEach((g, c) -> System.out.println("  " + g + ": " + c));
                }
                case "11" -> {
                    try {
                        collection.saveToFile("games_data.ser");
                        System.out.println("Сохранено.");
                    } catch (IOException e) {
                        System.out.println("Ошибка сохранения: " + e.getMessage());
                    }
                }
                case "12" -> {
                    try {
                        collection.loadFromFile("games_data.ser");
                        System.out.println("Загружено.");
                    } catch (IOException | ClassNotFoundException e) {
                        System.out.println("Ошибка загрузки: " + e.getMessage());
                    }
                }
                case "13" -> {
                    try {
                        collection.exportCSV("games_export.csv");
                        System.out.println("Экспортировано в games_export.csv");
                    } catch (IOException e) {
                        System.out.println("Ошибка экспорта: " + e.getMessage());
                    }
                }
                case "14" -> {
                    try {
                        collection.importCSV("games_export.csv");
                        System.out.println("Импортировано из games_export.csv");
                    } catch (IOException e) {
                        System.out.println("Ошибка импорта: " + e.getMessage());
                    }
                }
                default -> System.out.println("Неизвестная команда.");
            }
        }
    }
}
