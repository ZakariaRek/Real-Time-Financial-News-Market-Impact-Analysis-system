package org.example.src;

import javax.swing.*;
import java.awt.*;

public class GridGUI extends JFrame {
    private GridEnvironment environment;
    private JPanel gridPanel;
    private JLabel[][] cells;
    private JLabel statusLabel;
    private static final int CELL_SIZE = 80;

    public GridGUI(GridEnvironment env) {
        this.environment = env;
        environment.setGUI(this);

        setTitle("Système Multi-Agent JADE - Grille 6x6");
        setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
        setLayout(new BorderLayout());

        // Panel pour la grille
        gridPanel = new JPanel();
        gridPanel.setLayout(new GridLayout(6, 6, 2, 2));
        gridPanel.setBackground(Color.DARK_GRAY);
        gridPanel.setBorder(BorderFactory.createEmptyBorder(10, 10, 10, 10));

        // Initialiser les cellules
        cells = new JLabel[6][6];
        for (int i = 0; i < 6; i++) {
            for (int j = 0; j < 6; j++) {
                cells[i][j] = new JLabel();
                cells[i][j].setOpaque(true);
                cells[i][j].setHorizontalAlignment(SwingConstants.CENTER);
                cells[i][j].setVerticalAlignment(SwingConstants.CENTER);
                cells[i][j].setFont(new Font("Arial", Font.BOLD, 20));
                cells[i][j].setBorder(BorderFactory.createLineBorder(Color.BLACK, 2));
                cells[i][j].setPreferredSize(new Dimension(CELL_SIZE, CELL_SIZE));
                gridPanel.add(cells[i][j]);
            }
        }

        // Panel d'informations
        JPanel infoPanel = new JPanel();
        infoPanel.setLayout(new BorderLayout());
        infoPanel.setBackground(new Color(240, 240, 240));
        infoPanel.setBorder(BorderFactory.createEmptyBorder(10, 10, 10, 10));

        // Titre
        JLabel titleLabel = new JLabel("Navigation Agent JADE", SwingConstants.CENTER);
        titleLabel.setFont(new Font("Arial", Font.BOLD, 24));
        titleLabel.setForeground(new Color(0, 102, 204));

        // Légende
        JPanel legendPanel = new JPanel();
        legendPanel.setLayout(new FlowLayout(FlowLayout.CENTER, 20, 5));
        legendPanel.setBackground(new Color(240, 240, 240));

        legendPanel.add(createLegendItem("Départ", new Color(100, 200, 100)));
        legendPanel.add(createLegendItem("Agent", new Color(255, 200, 0)));
        legendPanel.add(createLegendItem("But", new Color(255, 100, 100)));
        legendPanel.add(createLegendItem("Vide", Color.WHITE));

        // Status
        statusLabel = new JLabel("Position Agent: (1, 1) → Objectif: (4, 5)", SwingConstants.CENTER);
        statusLabel.setFont(new Font("Arial", Font.PLAIN, 16));
        statusLabel.setBorder(BorderFactory.createEmptyBorder(10, 0, 0, 0));

        infoPanel.add(titleLabel, BorderLayout.NORTH);
        infoPanel.add(legendPanel, BorderLayout.CENTER);
        infoPanel.add(statusLabel, BorderLayout.SOUTH);

        add(infoPanel, BorderLayout.NORTH);
        add(gridPanel, BorderLayout.CENTER);

        updateDisplay();

        pack();
        setLocationRelativeTo(null);
        setVisible(true);
    }

    private JPanel createLegendItem(String text, Color color) {
        JPanel panel = new JPanel();
        panel.setLayout(new FlowLayout(FlowLayout.LEFT, 5, 0));
        panel.setBackground(new Color(240, 240, 240));

        JLabel colorBox = new JLabel("   ");
        colorBox.setOpaque(true);
        colorBox.setBackground(color);
        colorBox.setBorder(BorderFactory.createLineBorder(Color.BLACK, 1));

        JLabel label = new JLabel(text);
        label.setFont(new Font("Arial", Font.PLAIN, 14));

        panel.add(colorBox);
        panel.add(label);

        return panel;
    }

    public void updateDisplay() {
        SwingUtilities.invokeLater(() -> {
            int[][] grid = environment.getGrid();
            int agentX = environment.getAgentX();
            int agentY = environment.getAgentY();
            int startX = environment.getStartX();
            int startY = environment.getStartY();
            int goalX = environment.getGoalX();
            int goalY = environment.getGoalY();

            for (int i = 0; i < 6; i++) {
                for (int j = 0; j < 6; j++) {
                    if (i == agentX && j == agentY) {
                        // Position de l'agent
                        cells[i][j].setBackground(new Color(255, 200, 0));
                        cells[i][j].setText("🤖");
                        cells[i][j].setToolTipText("Agent (" + i + ", " + j + ")");
                    } else if (i == goalX && j == goalY) {
                        // Objectif
                        cells[i][j].setBackground(new Color(255, 100, 100));
                        cells[i][j].setText("🎯");
                        cells[i][j].setToolTipText("But (" + i + ", " + j + ")");
                    } else if (i == startX && j == startY) {
                        // Départ
                        cells[i][j].setBackground(new Color(100, 200, 100));
                        cells[i][j].setText("🏁");
                        cells[i][j].setToolTipText("Départ (" + i + ", " + j + ")");
                    } else {
                        // Cellule vide
                        cells[i][j].setBackground(Color.WHITE);
                        cells[i][j].setText("");
                        cells[i][j].setToolTipText("(" + i + ", " + j + ")");
                    }
                }
            }

            // Mettre à jour le status
            if (environment.isGoalReached()) {
                statusLabel.setText("✓ OBJECTIF ATTEINT! Agent a atteint (" + goalX + ", " + goalY + ")");
                statusLabel.setForeground(new Color(0, 150, 0));
                statusLabel.setFont(new Font("Arial", Font.BOLD, 18));
            } else {
                statusLabel.setText("Position Agent: (" + agentX + ", " + agentY + ") → Objectif: (" + goalX + ", " + goalY + ")");
                statusLabel.setForeground(Color.BLACK);
            }

            repaint();
        });
    }
}