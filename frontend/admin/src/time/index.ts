import { parse, differenceInMilliseconds } from 'date-fns';
import { format } from 'date-fns';
export {
  format,
  isValid,
  addSeconds,
  startOfMonth,
  endOfMonth,
} from 'date-fns';
import ptBR from 'date-fns/locale/pt-BR';
export const diff = differenceInMilliseconds;
import cronstrue from 'cronstrue';
import 'cronstrue/locales/pt_BR';

export const tformat = {
  SPTBR: 'dd/MM/yyyy',
  DASHUN: 'yyyy-MM-dd',
  EXTPTBR: "dd 'de' MMMM, yyyy",
  MONYEA: 'MM/yyyy',
  MONYEAR: 'MM/yy',
  DAYMON: 'dd/MM',
  RFC3339: `yyyy-MM-dd'T'HH:mm:ssXXX`,
  RFC3349: `yyyy-MM-dd'T'HH:mm:ss'Z'`,
  ISO8601: `yyyy-MM-dd'T'HH:mm:ss.SSSSSSX`,
  TIME: 'hh:mm',
  TIME24H: 'HH:mm',
  FULLTIME24H: 'HH:mm:ss',
  FULLDATE: 'dd/MM/yyyy HH:mm:ss',
  DAYMONYEAH: 'dd MMM yyyy HH:mm',
  FULLDAY: 'dddd',
  SHORTMONTH: 'MMM',
  MONTH: 'MM',
  YEAR: 'yy',
  VARIANTEXTPTBR: "dd 'de' MMMM 'de' yyyy",
  STPEUA: 'yyyy/MM/dd',
  SHORTSTP: 'YYY/MM/dd',
  DATEHOUR: 'yyyy-MM-dd HH:mm:ss',
  DATETIMELOCAL: `yyyy-MM-dd'T'HH:mm`,
  YEAMON: 'yyyy-MM',
  HUDDLED: 'yyyyMMdd',
  HUDDLEDTIME: 'HHmm',
  HUDDLEDDATEHOUR: 'yyyyMMddHHmm',
  SHRTPTBR: "dd 'de' MMMM",
};

export function RFC3339(time: string) {
  return parse(time, tformat.RFC3339, new Date());
}

export function MONYEAR(time: string) {
  return parse(time, tformat.MONYEAR, new Date());
}

export function ISO8601(time: string) {
  return parse(time, tformat.ISO8601, new Date());
}

export function dateFormatForDisplay(time: Date, to: string, def?: string) {
  try {
    return format(time, to);
  } catch (e) {
    return def ?? '--';
  }
}

export function dateFormatForDisplayPTBR(time: Date, to: string, def?: string) {
  try {
    return format(time, to, { locale: ptBR });
  } catch (e) {
    return def ?? '--';
  }
}

export function formatForDisplay(
  time: string,
  tformat: string,
  to: string,
  def?: string
) {
  try {
    const date = parse(time, tformat, new Date());
    return format(date, to);
  } catch (e) {
    return def ?? '--';
  }
}

export function readableCron(expr: string) {
  try {
    return cronstrue.toString(expr, { locale: 'pt_BR' });
  } catch (e) {
    return expr;
  }
}
