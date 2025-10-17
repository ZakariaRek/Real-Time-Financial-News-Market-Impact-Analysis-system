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
    private Random random;
    private int moveCount = 0;
    private Set<String> visitedCells;

    @Override
    protected void setup() {
        System.out.println("Agent Explorateur " + getLocalName() + " est prêt!");
        System.out.println("Position de départ: (" + currentX + ", " + currentY + ")");
        System.out.println("Position cible: (" + targetX + ", " + targetY + ")");
        System.out.println("Mode: Déplacement ALÉATOIRE jusqu'à trouver le but\n");

        random = new Random();
        visitedCells = new HashSet<>();
        visitedCells.add(currentX + "," + currentY);

        // Récupérer l'environnement depuis les arguments
        Object[] args = getArguments();
        if (args != null && args.length > 0) {
            grid = (GridEnvironment) args[0];
            grid.updateAgentPosition(currentX, currentY);
        }

        // Comportement pour se déplacer aléatoirement
        addBehaviour(new TickerBehaviour(this, 800) {
            @Override
            protected void onTick() {
                // Vérifier si l'objectif est atteint
                if (currentX == targetX && currentY == targetY) {
                    System.out.println("\n🎉 OBJECTIF ATTEINT! 🎉");
                    System.out.println("Position finale: (" + currentX + ", " + currentY + ")");
                    System.out.println("Nombre total de déplacements: " + moveCount);
                    System.out.println("Cellules visitées: " + visitedCells.size() + "/36");

                    // Informer l'environnement
                    if (grid != null) {
                        grid.setGoalReached(true);
                    }

                    // Arrêter le comportement
                    stop();
                    return;
                }

                // Se déplacer aléatoirement
                moveRandomly();
                moveCount++;
            }
        });
    }

    private void moveRandomly() {
        // Liste des mouvements possibles (haut, bas, gauche, droite)
        List<int[]> possibleMoves = new ArrayList<>();

        // Haut
        if (currentX > 0) {
            possibleMoves.add(new int[]{currentX - 1, currentY});
        }
        // Bas
        if (currentX < 5) {
            possibleMoves.add(new int[]{currentX + 1, currentY});
        }
        // Gauche
        if (currentY > 0) {
            possibleMoves.add(new int[]{currentX, currentY - 1});
        }
        // Droite
        if (currentY < 5) {
            possibleMoves.add(new int[]{currentX, currentY + 1});
        }

        // Choisir un mouvement aléatoire parmi les possibles
        if (!possibleMoves.isEmpty()) {
            int[] move = possibleMoves.get(random.nextInt(possibleMoves.size()));
            int newX = move[0];
            int newY = move[1];

            // Déplacer l'agent
            moveToPosition(newX, newY);

            // Ajouter à la liste des cellules visitées
            String cellKey = newX + "," + newY;
            boolean isNewCell = visitedCells.add(cellKey);

            // Afficher le mouvement
            String direction = getDirection(currentX, currentY, newX, newY);
            if (isNewCell) {
                System.out.println("Mouvement #" + moveCount + ": " + direction + " → (" + newX + ", " + newY + ") [Nouvelle cellule]");
            } else {
                System.out.println("Mouvement #" + moveCount + ": " + direction + " → (" + newX + ", " + newY + ") [Déjà visitée]");
            }

            // Vérifier si l'objectif est atteint
            if (newX == targetX && newY == targetY) {
                System.out.println("\n🎯 BUT TROUVÉ! 🎯");
            }
        }
    }

    private String getDirection(int oldX, int oldY, int newX, int newY) {
        if (newX < oldX) return "↑ HAUT";
        if (newX > oldX) return "↓ BAS";
        if (newY < oldY) return "← GAUCHE";
        if (newY > oldY) return "→ DROITE";
        return "•";
    }

    private void moveToPosition(int x, int y) {
        currentX = x;
        currentY = y;

        if (grid != null) {
            grid.updateAgentPosition(currentX, currentY);
        }
    }

    @Override
    protected void takeDown() {
        System.out.println("\n=== Statistiques finales ===");
        System.out.println("Agent " + getLocalName() + " se termine.");
        System.out.println("Total déplacements: " + moveCount);
        System.out.println("Cellules explorées: " + visitedCells.size() + "/36");
        System.out.println("===========================");
    }
}