package org.example.src;

public class GridEnvironment {
    private int[][] grid;
    private int agentX;
    private int agentY;
    private int startX = 1;
    private int startY = 1;
    private int goalX = 4;
    private int goalY = 5;
    private boolean goalReached = false;
    private GridGUI gui;

    public GridEnvironment() {
        grid = new int[6][6];
        agentX = startX;
        agentY = startY;
        initializeGrid();
    }

    private void initializeGrid() {
        // 0 = cellule vide
        // 1 = position de départ
        // 2 = position de l'agent
        // 3 = objectif
        for (int i = 0; i < 6; i++) {
            for (int j = 0; j < 6; j++) {
                grid[i][j] = 0;
            }
        }
        grid[startX][startY] = 1;
        grid[goalX][goalY] = 3;
    }

    public synchronized void updateAgentPosition(int x, int y) {
        // Effacer l'ancienne position (sauf si c'est le départ ou l'objectif)
        if (agentX != startX || agentY != startY) {
            if (agentX != goalX || agentY != goalY) {
                grid[agentX][agentY] = 0;
            }
        }

        // Mettre à jour la nouvelle position
        agentX = x;
        agentY = y;

        if (x != goalX || y != goalY) {
            grid[x][y] = 2;
        }

        // Notifier le GUI si présent
        if (gui != null) {
            gui.updateDisplay();
        }
    }

    public int[][] getGrid() {
        return grid;
    }

    public int getAgentX() {
        return agentX;
    }

    public int getAgentY() {
        return agentY;
    }

    public int getStartX() {
        return startX;
    }

    public int getStartY() {
        return startY;
    }

    public int getGoalX() {
        return goalX;
    }

    public int getGoalY() {
        return goalY;
    }

    public boolean isGoalReached() {
        return goalReached;
    }

    public void setGoalReached(boolean reached) {
        this.goalReached = reached;
        if (gui != null) {
            gui.updateDisplay();
        }
    }

    public void setGUI(GridGUI gui) {
        this.gui = gui;
    }

    public void printGrid() {
        System.out.println("\n=== État de la grille ===");
        for (int i = 0; i < 6; i++) {
            for (int j = 0; j < 6; j++) {
                if (i == agentX && j == agentY) {
                    System.out.print("[A] ");
                } else if (i == goalX && j == goalY) {
                    System.out.print("[G] ");
                } else if (i == startX && j == startY) {
                    System.out.print("[S] ");
                } else {
                    System.out.print("[ ] ");
                }
            }
            System.out.println();
        }
        System.out.println("=========================\n");
    }
}