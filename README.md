# 🪐 Gravity Trails: Soft-Body Climber

[![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Ebitengine](https://img.shields.io/badge/Ebitengine-v2.9.9-FB3E44?style=for-the-badge&logo=go)](https://ebitengine.org/)
[![GitHub Actions Deploy](https://img.shields.io/github/actions/workflow/status/o-semenov/gravity_trails/deploy.yml?branch=main&label=Deploy%20to%20GitHub%20Pages&style=for-the-badge)](https://github.com/o-semenov/gravity_trails/actions/workflows/deploy.yml)

> **Gravity Trails** — это двумерный симулятор преодоления препятствий с мягкой физикой тел (soft-body physics), разработанный на Go с использованием графического движка [Ebitengine](https://ebiten.org). Управляйте деформируемым багги или грузовиком, балансируйте в воздухе и преодолевайте крутые склоны!

🤖 **[Играть в браузере (GitHub Pages) ➔](https://o-semenov.github.io/gravity_trails/)**

---

## 📸 Скриншоты геймплея
| Главное меню |                                                            Игровой процесс                                                             |                                                          Мягкая деформация рамы                                                          |
|:---:|:--------------------------------------------------------------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------------------------------------------------:|
| ![Главное меню](https://raw.githubusercontent.com/o-semenov/gravity_trails/39db90d3eba2d7570173d391d8ccff0db5b868b6/menu.png) | ![Игровой процесс](https://raw.githubusercontent.com/o-semenov/gravity_trails/39db90d3eba2d7570173d391d8ccff0db5b868b6/game_play1.png) | ![Деформация машины](https://raw.githubusercontent.com/o-semenov/gravity_trails/39db90d3eba2d7570173d391d8ccff0db5b868b6/game_play2.png) |

---

## ✨ Основные особенности

* **Физика мягких тел (Soft-Body Physics):** Рама автомобиля состоит из узлов (Nodes) и пружин (Springs), что позволяет машине сжиматься, изгибаться и реалистично деформироваться при ударах.
* **Управление в воздухе (Air Control):** Балансируйте наклоном машины прямо в воздухе во время прыжков.
* **Интерактивное взаимодействие:** Захватывайте любые узлы автомобиля мышкой и перетаскивайте их прямо во время игры для преодоления сложных препятствий или восстановления формы!
* **Несколько типов транспорта:** Различные пресеты техники (Buggy, Truck) со своим уникальным весом, распределением центра масс и жесткостью пружин.
* **Система рекордов:** Игра сохраняет ваши лучшие достижения локально (`save.json`).

---

## 🛠️ Запуск

### Локальный запуск (Десктоп-версия)

Для запуска игры в нативном окне (требуется установленный Go SDK):

```bash
go run main.go
```
---