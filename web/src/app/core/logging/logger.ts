import { inject, Injectable } from '@angular/core';
//import { loggerConfig } from './logger.config';
import { environment } from "../../../environnements/environment";
import { LogLevel } from '../../../environnements/log-level';
import { HttpClient } from '@angular/common/http';


@Injectable({
  providedIn: 'root'
})
export class Logger {
  private logLevel = environment.logLevel;

  constructor(private http: HttpClient) {}

  private shouldLog(level: LogLevel): boolean {
    return level >= this.logLevel && this.logLevel !== LogLevel.OFF;
  }

  private logToConsole(level: LogLevel, message: string, ...params: any[]) {
    if (!this.shouldLog(level)) return;

    const timestamp = new Date().toISOString();
    const prefix = `[${timestamp}] [${LogLevel[level]}]`;

    switch (level) {
      case LogLevel.DEBUG:
        console.debug(prefix, message, ...params);
        break;
      case LogLevel.INFO:
        console.info(prefix, message, ...params);
        break;
      case LogLevel.WARN:
        console.warn(prefix, message, ...params);
        break;
      case LogLevel.ERROR:
        console.error(prefix, message, ...params);
        break;
      case LogLevel.FATAL:
        console.error(prefix, 'FATAL:', message, ...params);
        break;
    }
  }

  private logToServer(level: LogLevel, message: string, ...params: any[]) {
    if (environment.serverLoggingUrl && level >= LogLevel.ERROR) {
      this.http.post(environment.serverLoggingUrl, {
        level: LogLevel[level],
        message,
        params,
        app: environment.appName,
        env: environment.envName,
        timestamp: new Date()
      }).subscribe({
        error: err => console.error('Server logging failed', err)
      });
    }
  }

  debug(msg: string, ...params: any[]) {
    this.logToConsole(LogLevel.DEBUG, msg, ...params);
  }

  info(msg: string, ...params: any[]) {
    this.logToConsole(LogLevel.INFO, msg, ...params);
  }

  warn(msg: string, ...params: any[]) {
    this.logToConsole(LogLevel.WARN, msg, ...params);
  }

  error(msg: string, ...params: any[]) {
    this.logToConsole(LogLevel.ERROR, msg, ...params);
    this.logToServer(LogLevel.ERROR, msg, ...params);
  }

  fatal(msg: string, ...params: any[]) {
    this.logToConsole(LogLevel.FATAL, msg, ...params);
    this.logToServer(LogLevel.FATAL, msg, ...params);
  }
}