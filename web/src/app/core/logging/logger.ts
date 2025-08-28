import { inject, Injectable } from '@angular/core';
import { NgxLoggerLevel } from 'ngx-logger';
import { loggerConfig } from './logger.config';
import { NGXLogger } from 'ngx-logger';


import { environment } from "../../../environnements/environment";

@Injectable({
  providedIn: 'root'
})
export class Logger {
  private logger = inject(NGXLogger);

  constructor() {
    const logLevel = environment.production ? NgxLoggerLevel.INFO : NgxLoggerLevel.DEBUG;
    this.logger.updateConfig({ ...loggerConfig, level: logLevel });
  }

  debug(message: string, ...additional: any[]) {
    this.logger.debug(message, ...additional);
  }

  info(message: string, ...additional: any[]) {
    this.logger.info(message, ...additional);
  }

  warn(message: string, ...additional: any[]) {
    this.logger.warn(message, ...additional);
  }

  error(message: string, ...additional: any[]) {
    this.logger.error(message, ...additional);
  }

  fatal(message: string, ...additional: any[]) {
    this.logger.fatal(message, ...additional);
  }
}
