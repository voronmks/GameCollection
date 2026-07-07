# game_collection.rb
require 'json'
require 'date'
require 'csv'

class Game
  attr_accessor :id, :title, :genre, :platform, :year, :completed, :rating, :notes, :added_date

  def initialize(id, title, genre, platform, year, completed, rating, notes = "", added_date = Date.today.to_s)
    @id = id
    @title = title
    @genre = genre
    @platform = platform
    @year = year
    @completed = completed
    @rating = rating
    @notes = notes
    @added_date = added_date
  end

  def to_h
    { id: @id, title: @title, genre: @genre, platform: @platform, year: @year,
      completed: @completed, rating: @rating, notes: @notes, added_date: @added_date }
  end

  def self.from_h(hash)
    Game.new(hash[:id], hash[:title], hash[:genre], hash[:platform], hash[:year],
             hash[:completed], hash[:rating], hash[:notes], hash[:added_date])
  end
end

class GameCollection
  attr_reader :games

  def initialize
    @games = []
    @next_id = 1
  end

  def add_game(title, genre, platform, year, completed, rating, notes = "")
    raise "Оценка должна быть от 1 до 10" unless (1..10).include?(rating)
    raise "Год должен быть от 1980 до #{Date.today.year}" unless (1980..Date.today.year).include?(year)
    raise "Название, жанр и платформа не могут быть пустыми" if title.empty? || genre.empty? || platform.empty?
    game = Game.new(@next_id, title, genre, platform, year, completed, rating, notes)
    @games << game
    @next_id += 1
    game
  end

  def find_game(id)
    @games.find { |g| g.id == id }
  end

  def edit_game(id, **kwargs)
    game = find_game(id)
    return false unless game
    kwargs.each do |key, value|
      game.send("#{key}=", value) if game.respond_to?("#{key}=")
    end
    true
  end

  def delete_game(id)
    game = find_game(id)
    return false unless game
    @games.delete(game)
    true
  end

  def search_games(query)
    q = query.downcase
    @games.select { |g| g.title.downcase.include?(q) }
  end

  def filter_by_completed(completed)
    @games.select { |g| g.completed == completed }
  end

  def filter_by_genre(genre)
    @games.select { |g| g.genre.downcase == genre.downcase }
  end

  def filter_by_platform(platform)
    @games.select { |g| g.platform.downcase == platform.downcase }
  end

  def sort_by_rating(reverse = true)
    @games.sort_by { |g| g.rating }.reverse! if reverse
    @games.sort_by { |g| g.rating }
  end

  def sort_by_title
    @games.sort_by { |g| g.title.downcase }
  end

  def stats
    total = @games.size
    completed_count = filter_by_completed(true).size
    uncompleted = total - completed_count
    ratings = filter_by_completed(true).map(&:rating)
    avg_rating = ratings.empty? ? 0 : ratings.sum.to_f / ratings.size
    platforms = Hash.new(0)
    genres = Hash.new(0)
    @games.each do |g|
      platforms[g.platform] += 1
      genres[g.genre] += 1
    end
    { total: total, completed: completed_count, uncompleted: uncompleted,
      avg_rating: avg_rating, platforms: platforms, genres: genres }
  end

  def save_to_file(filename = "games_data.json")
    data = { games: @games.map(&:to_h) }
    File.write(filename, JSON.pretty_generate(data))
  end

  def load_from_file(filename = "games_data.json")
    return unless File.exist?(filename)
    data = JSON.parse(File.read(filename), symbolize_names: true)
    @games.clear
    data[:games].each do |item|
      game = Game.from_h(item)
      @games << game
      @next_id = game.id + 1 if game.id >= @next_id
    end
  rescue JSON::ParserError
    puts "Ошибка чтения файла."
  end

  def export_csv(filename = "games_export.csv")
    CSV.open(filename, "w", col_sep: ";") do |csv|
      csv << ["ID", "Название", "Жанр", "Платформа", "Год", "Пройдена", "Оценка", "Заметки", "Дата добавления"]
      @games.each do |g|
        csv << [g.id, g.title, g.genre, g.platform, g.year, g.completed ? "Да" : "Нет", g.rating, g.notes, g.added_date]
      end
    end
  end

  def import_csv(filename = "games_export.csv")
    unless File.exist?(filename)
      raise "Файл не найден"
    end
    CSV.foreach(filename, headers: true, col_sep: ";") do |row|
      begin
        add_game(
          title: row["Название"],
          genre: row["Жанр"],
          platform: row["Платформа"],
          year: row["Год"].to_i,
          completed: row["Пройдена"] == "Да",
          rating: row["Оценка"].to_i,
          notes: row["Заметки"]
        )
      rescue => e
        puts "Ошибка импорта строки: #{e}"
      end
    end
  end
end

def print_game(game)
  status = game.completed ? "✅ Пройдена" : "⏳ Не пройдена"
  puts "##{game.id} - #{game.title} (#{game.year})"
  puts "   Жанр: #{game.genre}, Платформа: #{game.platform}"
  puts "   #{status}, Оценка: #{game.rating}/10"
  puts "   Заметки: #{game.notes}" unless game.notes.empty?
  puts "   Добавлена: #{game.added_date}"
end

def main
  collection = GameCollection.new
  collection.load_from_file

  loop do
    puts "\n===== КОЛЛЕКЦИЯ ИГР (Ruby) ====="
    puts "1. Добавить игру"
    puts "2. Показать все игры"
    puts "3. Показать пройденные игры"
    puts "4. Показать непройденные игры"
    puts "5. Найти игры по названию"
    puts "6. Сортировать по оценке (по убыванию)"
    puts "7. Сортировать по названию"
    puts "8. Редактировать игру"
    puts "9. Удалить игру"
    puts "10. Показать статистику"
    puts "11. Сохранить в файл"
    puts "12. Загрузить из файла"
    puts "13. Экспорт в CSV"
    puts "14. Импорт из CSV"
    puts "0. Выход"
    print "Выберите действие: "
    choice = gets.chomp

    case choice
    when "0"
      break
    when "1"
      print "Название: "
      title = gets.chomp
      next if title.empty?
      print "Жанр: "
      genre = gets.chomp
      next if genre.empty?
      print "Платформа: "
      platform = gets.chomp
      next if platform.empty?
      print "Год выпуска: "
      year = gets.chomp.to_i
      print "Статус (1-пройдена, 0-нет): "
      completed = gets.chomp == "1"
      print "Оценка (1-10): "
      rating = gets.chomp.to_i
      print "Заметки (необязательно): "
      notes = gets.chomp
      begin
        game = collection.add_game(title, genre, platform, year, completed, rating, notes)
        puts "Игра добавлена с ID #{game.id}"
      rescue => e
        puts "Ошибка: #{e.message}"
      end
    when "2"
      if collection.games.empty?
        puts "Нет игр."
      else
        collection.games.each { |g| print_game(g) }
      end
    when "3"
      games = collection.filter_by_completed(true)
      if games.empty?
        puts "Нет пройденных игр."
      else
        games.each { |g| print_game(g) }
      end
    when "4"
      games = collection.filter_by_completed(false)
      if games.empty?
        puts "Нет непройденных игр."
      else
        games.each { |g| print_game(g) }
      end
    when "5"
      print "Введите часть названия: "
      query = gets.chomp
      results = collection.search_games(query)
      if results.empty?
        puts "Игры не найдены."
      else
        results.each { |g| print_game(g) }
      end
    when "6"
      sorted = collection.sort_by_rating(reverse: true)
      if sorted.empty?
        puts "Нет игр."
      else
        sorted.each { |g| print_game(g) }
      end
    when "7"
      sorted = collection.sort_by_title
      if sorted.empty?
        puts "Нет игр."
      else
        sorted.each { |g| print_game(g) }
      end
    when "8"
      print "Введите ID игры для редактирования: "
      id = gets.chomp.to_i
      game = collection.find_game(id)
      unless game
        puts "Игра не найдена."
        next
      end
      puts "Оставьте поле пустым, чтобы не менять."
      print "Название (#{game.title}): "
      new_title = gets.chomp
      print "Жанр (#{game.genre}): "
      new_genre = gets.chomp
      print "Платформа (#{game.platform}): "
      new_platform = gets.chomp
      print "Год (#{game.year}): "
      new_year = gets.chomp
      print "Статус (1-пройдена, 0-нет) сейчас: #{game.completed ? '1' : '0'}: "
      new_completed = gets.chomp
      print "Оценка (#{game.rating}): "
      new_rating = gets.chomp
      print "Заметки (#{game.notes}): "
      new_notes = gets.chomp
      updates = {}
      updates[:title] = new_title unless new_title.empty?
      updates[:genre] = new_genre unless new_genre.empty?
      updates[:platform] = new_platform unless new_platform.empty?
      unless new_year.empty?
        updates[:year] = new_year.to_i
      end
      unless new_completed.empty?
        updates[:completed] = new_completed == "1"
      end
      unless new_rating.empty?
        updates[:rating] = new_rating.to_i
      end
      updates[:notes] = new_notes unless new_notes.empty?
      if collection.edit_game(id, **updates)
        puts "Игра обновлена."
      else
        puts "Ошибка обновления."
      end
    when "9"
      print "Введите ID игры для удаления: "
      id = gets.chomp.to_i
      if collection.delete_game(id)
        puts "Игра удалена."
      else
        puts "Игра не найдена."
      end
    when "10"
      stats = collection.stats
      puts "\n=== СТАТИСТИКА ==="
      puts "Всего игр: #{stats[:total]}"
      puts "Пройдено: #{stats[:completed]}"
      puts "Не пройдено: #{stats[:uncompleted]}"
      puts "Средняя оценка (только пройденные): #{'%.2f' % stats[:avg_rating]}"
      puts "По платформам:"
      stats[:platforms].each { |p, c| puts "  #{p}: #{c}" }
      puts "По жанрам:"
      stats[:genres].each { |g, c| puts "  #{g}: #{c}" }
    when "11"
      collection.save_to_file
      puts "Сохранено."
    when "12"
      collection.load_from_file
      puts "Загружено."
    when "13"
      collection.export_csv
      puts "Экспортировано в games_export.csv"
    when "14"
      begin
        collection.import_csv
        puts "Импортировано из games_export.csv"
      rescue => e
        puts "Ошибка импорта: #{e}"
      end
    else
      puts "Неизвестная команда."
    end
  end
end

main if __FILE__ == $0
