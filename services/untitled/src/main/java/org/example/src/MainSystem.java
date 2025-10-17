package org.example.src;

import jade.core.Profile;
import jade.core.ProfileImpl;
import jade.core.Runtime;
import jade.wrapper.AgentContainer;
import jade.wrapper.AgentController;
import jade.wrapper.StaleProxyException;

public class MainSystem {

    public static void main(String[] args) {
        try {
            // Créer l'environnement
            GridEnvironment environment = new GridEnvironment();

            // Créer l'interface graphique
            GridGUI gui = new GridGUI(environment);

            // Configurer JADE
            Runtime rt = Runtime.instance();
            Profile profile = new ProfileImpl();
            profile.setParameter(Profile.MAIN_HOST, "localhost");
            profile.setParameter(Profile.GUI, "true"); // Activer le GUI JADE

            // Créer le conteneur principal
            AgentContainer mainContainer = rt.createMainContainer(profile);

            System.out.println("=== Système Multi-Agent JADE ===");
            System.out.println("Conteneur principal créé");
            System.out.println("Grille: 6x6");
            System.out.println("Position de départ: (1, 1)");
            System.out.println("Position but: (4, 5)");
            System.out.println("================================\n");

            // Créer et démarrer l'agent explorateur
            Object[] agentArgs = new Object[]{environment};
            AgentController explorerAgent = mainContainer.createNewAgent(
                    "Explorer",
                    "ExplorerAgent",
                    agentArgs
            );
            explorerAgent.start();

            System.out.println("Agent Explorateur démarré!");
            System.out.println("L'agent va naviguer de (1,1) vers (4,5)...\n");

            // Afficher l'état initial de la grille en console
            environment.printGrid();

        } catch (StaleProxyException e) {
            System.err.println("Erreur lors de la création des agents: " + e.getMessage());
            e.printStackTrace();
        } catch (Exception e) {
            System.err.println("Erreur: " + e.getMessage());
            e.printStackTrace();
        }
    }
}