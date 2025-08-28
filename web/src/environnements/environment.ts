import { LogLevel } from './log-level';

export const environment = {
    production: false,
    logLevel: LogLevel.DEBUG,
    serverLoggingUrl: '/api/logs',
    appName: 'immogestion',
    envName: 'dev',
    apiUrl: 'http://localhost:8080/api',
};