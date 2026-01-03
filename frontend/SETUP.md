# Установка зависимостей

✅ **Зависимости уже установлены!**

Если нужно переустановить:

```bash
npm install vue-router@4 pinia vue-i18n@9
```

Или установить все зависимости сразу:

```bash
npm install
```

## Зависимости проекта

### Основные зависимости:
- `vue` - Vue 3 фреймворк
- `vue-router@4` - Роутинг для Vue
- `pinia` - Управление состоянием
- `vue-i18n@9` - Многоязычность

### Dev зависимости:
- `typescript` - TypeScript
- `vite` - Сборщик
- `vue-tsc` - TypeScript проверка для Vue
- `@vitejs/plugin-vue` - Плагин Vue для Vite

## Запуск проекта

```bash
npm run dev
```

Проект будет доступен по адресу `http://localhost:5173`

## Структура проекта

```
frontend/
├── src/
│   ├── components/
│   │   ├── layout/          # Header, Footer, Navigation
│   │   ├── common/          # LanguageSelector, StepIndicator
│   │   ├── map-pool/        # MapPoolCard
│   │   └── veto/            # Существующие компоненты вето
│   ├── pages/               # Все страницы приложения
│   ├── composables/          # useVeto, useTimer, useI18n
│   ├── store/               # Pinia stores (auth, veto)
│   ├── router/              # Vue Router конфигурация
│   ├── locales/             # Переводы (en.json, ru.json)
│   └── types/                # TypeScript типы
└── public/
    └── images/              # Изображения карт
```

## Выполненные задачи (MVP)

✅ Задача 1: Настройка базовой инфраструктуры
✅ Задача 2: Компонент Header с навигацией
✅ Задача 3: Компонент Footer
✅ Задача 4: Страница выбора пула карт
✅ Задача 5: Интеграция процесса вето с роутингом
✅ Задача 6: Настройка многоязычности (i18n)
