export interface MyGeoJSONFeature {
    type: 'Feature';
    geometry: {
      type: 'Polygon' | 'MultiPolygon';
      coordinates: any; // or number[][][] if you want to be strict
    };
    properties: {
      [key: string]: any; 
    };
  }
  
  export interface AffectedArea {
    id: number;
    en: string;
    sv: string;
  }
  
  export interface WarningLevel {
    en: string;
    sv: string;
    code: string;
  }
  
  export interface Event {
    en: string;
    sv: string;
    code: string;
  }
  
  export interface Object {
    type: string;
    url: string;
  }
  
  export interface WarningArea {
    id: number;
    approximateStart?: string;
    approximateEnd?: string;
    area: MyGeoJSONFeature; // <-- updated
    affectedAreas: AffectedArea[];
    warningLevel: WarningLevel;
    eventDescription?: Event;
  }
  
  // The main warning object
  export interface Warning {
    id: number;
    event: Event;
    warningAreas?: WarningArea[];
    objects?: Object[];
  }
  