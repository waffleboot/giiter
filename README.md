- Строит из коммитов feature ветки другие review ветки вида `/review/feature/N`, по диффам которых потом можно создавать MR
- Перенаправляет ветки на нужные коммиты если они не менялись, удаляет устаревшие review ветки
- Для сопоставления коммитов использует diff hash
- Старая review ветка не удаляется если есть хоть один новый коммит чтобы не потерять MR

### Показать список коммитов feature ветки

```bash
$ giiter git l -b master -f feature

Плюсом отмечены новые коммиты
Точкой отмечены учтенные коммиты
Минусом отмечены устаревшие review ветки

1) . 3694811 [review/feature/1] 1
2) + c205c0d 222
3) . 7381ab8 [review/feature/3] 3
4) . 554fac5 [review/feature/4] 4
5) . 45af91b [review/feature/5] 5
6) - 4b15cbd [review/feature/2] 222
```

### Пересадить коммит на старую review ветку чтобы не потерять MR

```bash
$ giiter git a -b master -f feature 2 6

2 и 6 это порядковый номер коммита и порядковый номер ветки

1) . 3694811 [review/feature/1] 1
2) . c205c0d [review/feature/2] 222
3) . 7381ab8 [review/feature/3] 3
4) . 554fac5 [review/feature/4] 4
5) . 45af91b [review/feature/5] 5
```

### Создать review ветки и gitlab MR

```bash
$ giiter git m -b master -f feature
Точками отмечены коммиты которые есть в MR и review ветках
1) . 3694811 [review/feature/1] 1
2) . e652189 [review/feature/2] 222
3) . 0d20232 [review/feature/3] 3
4) . ad3826b [review/feature/4] 4
5) . 295e75d [review/feature/5] 5
```

### Удалить review ветки

```bash
$ giiter git d -f feature
```