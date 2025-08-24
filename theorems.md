## THMS For minesweeper
- Effective Number: EN := cell.number - cell.numAdjacentFlags()
- `grid.mineConfigurations(i, j) ->` `[][]Pos`:
```
    let nbs: []Cell := cell object of each neighbor
    let unflaggedNbs := nbs - (pos of each flagged nb)

    let n := len(nbs)
    let EN := grid(i, j).number - grid(i, j).numAdjacentFlags() 

    return all possible configurations of surrounding mines such that
    the EN reaches 0, and doesn't violate any configurations of its own neighbors
```
        

