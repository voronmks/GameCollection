// game_collection.js
const fs = require('fs').promises;
const readline = require('readline');
const csv = require('csv-parser');
const createCsvWriter = require('csv-writer').createObjectCsvWriter;
const { parse } = require('csv-parse');

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

const question = (prompt) => new Promise(resolve => rl.question(prompt, resolve));

class Game {
    constructor(id, title, genre, platform, year, completed, rating, notes, addedDate) {
        this.id = id;
        this.title = title;
        this.genre = genre;
        this.platform = platform;
        this.year = year;
        this.completed = completed;
        this.rating = rating;
        this.notes = notes || '';
        this.addedDate = addedDate || new Date().toISOString().slice(0, 10);
    }
}

class GameCollection {
    constructor() {
        this.games = [];
        this.nextId = 1;
    }

    addGame(title, genre, platform, year, completed, rating, notes) {
        if (rating < 1 || rating > 10) throw new Error('Оценка должна быть от 1 до 10');
        const currentYear = new Date().getFullYear();
        if (year < 1980 || year > currentYear) throw new Error(`Год должен быть от 1980 до ${currentYear}`);
        if (!title.trim() || !genre.trim() || !platform.trim())
            throw new Error('Название, жанр и платформа не могут быть пустыми');
        const game = new Game(this.nextId, title, genre, platform, year, completed, rating, notes);
        this.games.push(game);
        this.nextId++;
        return game;
    }

    findGame(id) {
        return this.games.find(g => g.id === id);
    }

    editGame(id, updates) {
        const game = this.findGame(id);
        if (!game) return false;
        Object.assign(game, updates);
        return true;
    }

    deleteGame(id) {
        const index = this.games.findIndex(g => g.id === id);
        if (index === -1) return false;
        this.games.splice(index, 1);
        return true;
    }

    searchGames(query) {
        const q = query.toLowerCase();
        return this.games.filter(g => g.title.toLowerCase().includes(q));
    }

    filterByCompleted(completed) {
        return this.games.filter(g => g.completed === completed);
    }

    filterByGenre(genre) {
        return this.games.filter(g => g.genre.toLowerCase() === genre.toLowerCase());
    }

    filterByPlatform(platform) {
        return this.games.filter(g => g.platform.toLowerCase() === platform.toLowerCase());
    }

    sortByRating(reverse = true) {
        const sorted = [...this.games];
        sorted.sort((a, b) => reverse ? b.rating - a.rating : a.rating - b.rating);
        return sorted;
    }

    sortByTitle() {
        return [...this.games].sort((a, b) => a.title.localeCompare(b.title, undefined, { sensitivity: 'base' }));
    }

    getStats() {
        const total = this.games.length;
        const completedCount = this.filterByCompleted(true).length;
        const uncompleted = total - completedCount;
        const ratings = this.filterByCompleted(true).map(g => g.rating);
        const avgRating = ratings.length ? ratings.reduce((a, b) => a + b, 0) / ratings.length : 0;
        const platforms = {};
        const genres = {};
        this.games.forEach(g => {
            platforms[g.platform] = (platforms[g.platform] || 0) + 1;
            genres[g.genre] = (genres[g.genre] || 0) + 1;
        });
        return { total, completed: completedCount, uncompleted, avgRating, platforms, genres };
    }

    async saveToFile(filename = 'games_data.json') {
        const data = { games: this.games };
        await fs.writeFile(filename, JSON.stringify(data, null, 2));
    }

    async loadFromFile(filename = 'games_data.json') {
        try {
            const data = await fs.readFile(filename, 'utf8');
            const parsed = JSON.parse(data);
            this.games = parsed.games.map(g => Object.assign(new Game(0), g));
            this.nextId = this.games.reduce((max, g) => Math.max(max, g.id), 0) + 1;
        } catch (err) {
            if (err.code !== 'ENOENT') throw err;
        }
    }

    async exportCSV(filename = 'games_export.csv') {
        const records = this.games.map(g => ({
            ID: g.id,
            Название: g.title,
            Жанр: g.genre,
            Платформа: g.platform,
            Год: g.year,
            Пройдена: g.completed ? 'Да' : 'Нет',
            Оценка: g.rating,
            Заметки: g.notes,
            'Дата добавления': g.addedDate
        }));
        const csvWriter = createCsvWriter({
            path: filename,
            header: [
                { id: 'ID', title: 'ID' },
                { id: 'Название', title: 'Название' },
                { id: 'Жанр', title: 'Жанр' },
                { id: 'Платформа', title: 'Платформа' },
                { id: 'Год', title: 'Год' },
                { id: 'Пройдена', title: 'Пройдена' },
                { id: 'Оценка', title: 'Оценка' },
                { id: 'Заметки', title: 'Заметки' },
                { id: 'Дата добавления', title: 'Дата добавления' }
            ],
            fieldDelimiter: ';'
        });
        await csvWriter.writeRecords(records);
    }

    async importCSV(filename = 'games_export.csv') {
        const fileContent = await fs.readFile(filename, 'utf8');
        return new Promise((resolve, reject) => {
            parse(fileContent, { columns: true, delimiter: ';' }, (err, records) => {
                if (err) reject(err);
                for (const row of records) {
                    try {
                        this.addGame(
                            row['Название'],
                            row['Жанр'],
                            row['Платформа'],
                            parseInt(row['Год']),
                            row['Пройдена'] === 'Да',
                            parseInt(row['Оценка']),
                            row['Заметки']
                        );
                    } catch (e) {
                        console.log('Ошибка импорта строки:', e.message);
                    }
                }
                resolve();
            });
        });
    }
}

function printGame(game) {
    const status = game.completed ? '✅ Пройдена' : '⏳ Не пройдена';
    console.log(`#${game.id} - ${game.title} (${game.year})`);
    console.log(`   Жанр: ${game.genre}, Платформа: ${game.platform}`);
    console.log(`   ${status}, Оценка: ${game.rating}/10`);
    if (game.notes) console.log(`   Заметки: ${game.notes}`);
    console.log(`   Добавлена: ${game.addedDate}`);
}

async function main() {
    const collection = new GameCollection();
    await collection.loadFromFile();

    while (true) {
        console.log('\n===== КОЛЛЕКЦИЯ ИГР (JavaScript) =====');
        console.log('1. Добавить игру');
        console.log('2. Показать все игры');
        console.log('3. Показать пройденные игры');
        console.log('4. Показать непройденные игры');
        console.log('5. Найти игры по названию');
        console.log('6. Сортировать по оценке (по убыванию)');
        console.log('7. Сортировать по названию');
        console.log('8. Редактировать игру');
        console.log('9. Удалить игру');
        console.log('10. Показать статистику');
        console.log('11. Сохранить в файл');
        console.log('12. Загрузить из файла');
        console.log('13. Экспорт в CSV');
        console.log('14. Импорт из CSV');
        console.log('0. Выход');
        const choice = await question('Выберите действие: ');

        if (choice === '0') break;

        switch (choice) {
            case '1': {
                const title = await question('Название: ');
                if (!title.trim()) { console.log('Название не может быть пустым.'); continue; }
                const genre = await question('Жанр: ');
                if (!genre.trim()) { console.log('Жанр не может быть пустым.'); continue; }
                const platform = await question('Платформа: ');
                if (!platform.trim()) { console.log('Платформа не может быть пустой.'); continue; }
                const year = parseInt(await question('Год выпуска: '));
                const completed = (await question('Статус (1-пройдена, 0-нет): ')) === '1';
                const rating = parseInt(await question('Оценка (1-10): '));
                const notes = await question('Заметки (необязательно): ');
                try {
                    const game = collection.addGame(title, genre, platform, year, completed, rating, notes);
                    console.log(`Игра добавлена с ID ${game.id}`);
                } catch (err) {
                    console.log('Ошибка:', err.message);
                }
                break;
            }
            case '2':
                if (collection.games.length === 0) console.log('Нет игр.');
                else collection.games.forEach(printGame);
                break;
            case '3': {
                const games = collection.filterByCompleted(true);
                if (games.length === 0) console.log('Нет пройденных игр.');
                else games.forEach(printGame);
                break;
            }
            case '4': {
                const games = collection.filterByCompleted(false);
                if (games.length === 0) console.log('Нет непройденных игр.');
                else games.forEach(printGame);
                break;
            }
            case '5': {
                const query = await question('Введите часть названия: ');
                const results = collection.searchGames(query);
                if (results.length === 0) console.log('Игры не найдены.');
                else results.forEach(printGame);
                break;
            }
            case '6': {
                const sorted = collection.sortByRating(true);
                if (sorted.length === 0) console.log('Нет игр.');
                else sorted.forEach(printGame);
                break;
            }
            case '7': {
                const sorted = collection.sortByTitle();
                if (sorted.length === 0) console.log('Нет игр.');
                else sorted.forEach(printGame);
                break;
            }
            case '8': {
                const id = parseInt(await question('Введите ID игры для редактирования: '));
                const game = collection.findGame(id);
                if (!game) { console.log('Игра не найдена.'); continue; }
                console.log('Оставьте поле пустым, чтобы не менять.');
                const newTitle = await question(`Название (${game.title}): `);
                const newGenre = await question(`Жанр (${game.genre}): `);
                const newPlatform = await question(`Платформа (${game.platform}): `);
                const newYear = await question(`Год (${game.year}): `);
                const newCompleted = await question(`Статус (1-пройдена, 0-нет) сейчас: ${game.completed ? '1' : '0'}: `);
                const newRating = await question(`Оценка (${game.rating}): `);
                const newNotes = await question(`Заметки (${game.notes}): `);
                const updates = {};
                if (newTitle.trim()) updates.title = newTitle;
                if (newGenre.trim()) updates.genre = newGenre;
                if (newPlatform.trim()) updates.platform = newPlatform;
                if (newYear.trim()) {
                    const y = parseInt(newYear);
                    if (!isNaN(y)) updates.year = y;
                    else console.log('Год должен быть числом, пропускаем.');
                }
                if (newCompleted.trim()) updates.completed = newCompleted === '1';
                if (newRating.trim()) {
                    const r = parseInt(newRating);
                    if (!isNaN(r)) updates.rating = r;
                    else console.log('Оценка должна быть числом, пропускаем.');
                }
                if (newNotes.trim()) updates.notes = newNotes;
                if (collection.editGame(id, updates)) console.log('Игра обновлена.');
                else console.log('Ошибка обновления.');
                break;
            }
            case '9': {
                const id = parseInt(await question('Введите ID игры для удаления: '));
                if (collection.deleteGame(id)) console.log('Игра удалена.');
                else console.log('Игра не найдена.');
                break;
            }
            case '10': {
                const stats = collection.getStats();
                console.log('\n=== СТАТИСТИКА ===');
                console.log(`Всего игр: ${stats.total}`);
                console.log(`Пройдено: ${stats.completed}`);
                console.log(`Не пройдено: ${stats.uncompleted}`);
                console.log(`Средняя оценка (только пройденные): ${stats.avgRating.toFixed(2)}`);
                console.log('По платформам:');
                for (const [p, c] of Object.entries(stats.platforms)) {
                    console.log(`  ${p}: ${c}`);
                }
                console.log('По жанрам:');
                for (const [g, c] of Object.entries(stats.genres)) {
                    console.log(`  ${g}: ${c}`);
                }
                break;
            }
            case '11':
                try {
                    await collection.saveToFile();
                    console.log('Сохранено.');
                } catch (err) {
                    console.log('Ошибка сохранения:', err.message);
                }
                break;
            case '12':
                try {
                    await collection.loadFromFile();
                    console.log('Загружено.');
                } catch (err) {
                    console.log('Ошибка загрузки:', err.message);
                }
                break;
            case '13':
                try {
                    await collection.exportCSV();
                    console.log('Экспортировано в games_export.csv');
                } catch (err) {
                    console.log('Ошибка экспорта:', err.message);
                }
                break;
            case '14':
                try {
                    await collection.importCSV();
                    console.log('Импортировано из games_export.csv');
                } catch (err) {
                    console.log('Ошибка импорта:', err.message);
                }
                break;
            default:
                console.log('Неизвестная команда.');
        }
    }
    rl.close();
}

main().catch(console.error);
