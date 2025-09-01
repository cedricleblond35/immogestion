import { LogLevel } from './log-level';

export const environment = {
    production: false,
    logLevel: LogLevel.DEBUG,
    serverLoggingUrl: '/api/logs',
    appName: 'immogestion',
    envName: 'dev',
    apiUrl: 'http://localhost:8080/api',
    apiVersion: 'v1',
    timeout: 10000,
    retryAttempts: 2,
    TOKEN_KEY: 'immogestion_access_token',
    REFRESH_TOKEN_KEY: 'immogestion_refresh_token',
    USER_KEY: 'immogestion_user'
};