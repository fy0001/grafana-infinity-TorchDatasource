import { forEach, get, toNumber } from 'lodash';
import { parseString } from 'xml2js';
import { InfinityParser } from './InfinityParser';
import { InfinityQuery, ScrapColumn, GrafanaTableRow } from './../../types';

export class XMLParser extends InfinityParser {
  constructor(XMLResponse: any | string, target: InfinityQuery, endTime?: Date) {
    super(target);
    this.formatInput(XMLResponse).then((xmlResponse: any) => {
      if (this.target.root_selector) {
        xmlResponse = get(xmlResponse, this.target.root_selector);
      }
      if (Array.isArray(xmlResponse)) {
        this.constructTableData(xmlResponse);
        this.constructTimeSeriesData(xmlResponse, endTime);
      } else {
        this.constructSingleTableData(xmlResponse);
      }
    });
  }
  private formatInput(XMLResponse: string) {
    return new Promise((resolve, reject) => {
      parseString(XMLResponse, (err, res) => {
        resolve(res);
      });
    });
  }
  private constructTableData(XMLResponse: any[]) {
    forEach(XMLResponse, r => {
      const row: GrafanaTableRow = [];
      this.target.columns.forEach((c: ScrapColumn) => {
        let value = get(r, c.selector, '');
        if (c.type === 'timestamp') {
          value = new Date(value + '');
        } else if (c.type === 'timestamp_epoch') {
          value = new Date(parseInt(value, 10));
        } else if (c.type === 'number') {
          value = value === '' ? null : +value;
        }
        if (typeof r === 'string') {
          row.push(r);
        } else {
          row.push(value);
        }
      });
      this.rows.push(row);
    });
  }
  private constructTimeSeriesData(XMLResponse: object, endTime: Date | undefined) {
    this.NumbersColumns.forEach((metricColumn: ScrapColumn) => {
      forEach(XMLResponse, r => {
        let seriesName = this.StringColumns.map(c => r[c.selector]).join(' ');
        if (this.NumbersColumns.length > 1) {
          seriesName += ` ${metricColumn.text}`;
        }
        if (this.NumbersColumns.length === 1 && seriesName === '') {
          seriesName = `${metricColumn.text}`;
        }
        seriesName = seriesName.trim();
        let timestamp = endTime ? endTime.getTime() : new Date().getTime();
        if (this.TimeColumns.length >= 1) {
          const FirstTimeColumn = this.TimeColumns[0];
          if (FirstTimeColumn.type === 'timestamp') {
            timestamp = new Date(get(r, FirstTimeColumn.selector) + '').getTime();
          } else if (FirstTimeColumn.type === 'timestamp_epoch') {
            timestamp = new Date(parseInt(get(r, FirstTimeColumn.selector), 10)).getTime();
          }
        }
        let metric = toNumber(get(r, metricColumn.selector));
        this.series.push({
          target: seriesName,
          datapoints: [[metric, timestamp]],
        });
      });
    });
  }
  private constructSingleTableData(XMLResponse: object) {
    const row: GrafanaTableRow = [];
    this.target.columns.forEach((c: ScrapColumn) => {
      row.push(get(XMLResponse, c.selector, ''));
    });
    this.rows.push(row);
  }
}
