import { NgxLoggerLevel } from "ngx-logger";
import { environment } from "../../../environnements/environment";

export const loggerConfig: Partial<any> = {
    level: NgxLoggerLevel.DEBUG,  // Niveau par défaut (DEBUG en dev, INFO ou ERROR en prod )
    serverLogLevel: NgxLoggerLevel.ERROR,// Niveau pour les logs envoyés au serveur
    disableConsoleLogging: false, // Désactiver la console en prod
    enableSourceMaps: true,   // Activer les source maps pour le débogage
    timestampFormat: 'HH:mm:ss.SSS',
    serverLoggingUrl: '/api/logs',    // URL pour envoyer les logs au serveur 
    colorScheme: [
    'purple',    // DEBUG
    'teal',      // INFO
    'gray',      // LOG
    'gray',      // WARN
    'red',       // ERROR
    'maroon',    // FATAL
    'black',     // OFF
  ],
    maxLogLength: 1000, // Longueur maximale d'un message de log
    prettifyJson: true, // Formater les objets JSON pour une meilleure lisibilité
    disableFileDetails: true, // Afficher les détails du fichier (nom, ligne) dans les logs
    additionalFields: {
    appName: 'immogestion',
    environment: environment.envName,
  },
};