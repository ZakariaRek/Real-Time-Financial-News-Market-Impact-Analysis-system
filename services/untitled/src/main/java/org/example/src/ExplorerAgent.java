package org.example.src;

import jade.core.Agent;
import jade.core.behaviours.TickerBehaviour;
import jade.lang.acl.ACLMessage;
import java.util.*;

public class ExplorerAgent extends Agent {
    private int currentX = 1;
    private int currentY = 1;
    private int targetX = 4;
    private int targetY = 5;
    private GridEnvironment grid;
    private List<Point> path;
    private int pathIndex = 0;

    class Point {
        int x, y;
        Point(int x, int y) {
            this.x = x;
            this.y = y;
        }
    }

    @Override
    protected void setup() {
        System.out.println("Agent Explorateur " + getLocalName() + " est prêt!");
        System.out.println("Position de départ: (" + currentX + ", " + currentY + ")");
        System.out.println("Position cible: (" + targetX + ", " + targetY + ")");

        // Récupérer l'environnement depuis les arguments
        Object[] args = getArguments();
        if (args != null && args.length > 0) {
            grid = (GridEnvironment) args[0];
            grid.updateAgentPosition(currentX, currentY);
        }

        // Calculer le chemin avec A*
        path = findPath();

        // Comportement pour se déplacer
        addBehaviour(new TickerBehaviour(this, 1000) {
            @Override
            protected void onTick() {
                if (pathIndex < path.size()) {
                    Point nextPoint = path.get(pathIndex);
                    moveToPosition(nextPoint.x, nextPoint.y);
                    pathIndex++;

                    if (currentX == targetX && currentY == targetY) {
                        System.out.println("Objectif atteint! Position: (" + currentX + ", " + currentY + ")");
                        // Informer l'environnement
                        if (grid != null) {
                            grid.setGoalReached(true);
                        }
                    }
                } else if (currentX == targetX && currentY == targetY) {
                    // Objectif atteint, arrêter
                    stop();
                }
            }
        });
    }

    private List<Point> findPath() {
        // Algorithme A* pour trouver le chemin
        PriorityQueue<Node> openSet = new PriorityQueue<>((a, b) ->
                Double.compare(a.f, b.f));
        Set<String> closedSet = new HashSet<>();
        Map<String, Node> allNodes = new HashMap<>();

        Node startNode = new Node(currentX, currentY, null);
        startNode.g = 0;
        startNode.h = heuristic(currentX, currentY);
        startNode.f = startNode.h;

        openSet.add(startNode);
        allNodes.put(currentX + "," + currentY, startNode);

        while (!openSet.isEmpty()) {
            Node current = openSet.poll();

            if (current.x == targetX && current.y == targetY) {
                return reconstructPath(current);
            }

            closedSet.add(current.x + "," + current.y);

            // Vérifier les 4 directions
            int[][] directions = {{0, 1}, {1, 0}, {0, -1}, {-1, 0}};
            for (int[] dir : directions) {
                int newX = current.x + dir[0];
                int newY = current.y + dir[1];

                if (newX < 0 || newX >= 6 || newY < 0 || newY >= 6) {
                    continue;
                }

                String key = newX + "," + newY;
                if (closedSet.contains(key)) {
                    continue;
                }

                double tentativeG = current.g + 1;

                Node neighbor = allNodes.get(key);
                if (neighbor == null) {
                    neighbor = new Node(newX, newY, current);
                    neighbor.g = tentativeG;
                    neighbor.h = heuristic(newX, newY);
                    neighbor.f = neighbor.g + neighbor.h;
                    allNodes.put(key, neighbor);
                    openSet.add(neighbor);
                } else if (tentativeG < neighbor.g) {
                    neighbor.parent = current;
                    neighbor.g = tentativeG;
                    neighbor.f = neighbor.g + neighbor.h;
                    openSet.remove(neighbor);
                    openSet.add(neighbor);
                }
            }
        }

        // Si aucun chemin trouvé, retourner chemin vide
        return new ArrayList<>();
    }

    private double heuristic(int x, int y) {
        return Math.abs(x - targetX) + Math.abs(y - targetY);
    }

    private List<Point> reconstructPath(Node node) {
        List<Point> path = new ArrayList<>();
        Node current = node;
        while (current != null) {
            path.add(0, new Point(current.x, current.y));
            current = current.parent;
        }
        return path;
    }

    private void moveToPosition(int x, int y) {
        currentX = x;
        currentY = y;
        System.out.println("Déplacement vers: (" + currentX + ", " + currentY + ")");

        if (grid != null) {
            grid.updateAgentPosition(currentX, currentY);
        }
    }

    class Node {
        int x, y;
        Node parent;
        double g, h, f;

        Node(int x, int y, Node parent) {
            this.x = x;
            this.y = y;
            this.parent = parent;
        }
    }

    @Override
    protected void takeDown() {
        System.out.println("Agent " + getLocalName() + " se termine.");
    }
}