// src/components/SwedenMap.tsx
import React, { useRef, useEffect } from 'react';
import {
  MapContainer,
  WMSTileLayer,
  GeoJSON,
  CircleMarker,
  Tooltip,
  MapContainerProps,
} from 'react-leaflet';
import L from 'leaflet';
import type { Feature, FeatureCollection } from 'geojson';
import { WarningArea } from '../Types/WarningTypes';

interface SwedenMapProps {
  warningAreas: WarningArea[];
  sensors: Sensor[];
}

interface Sensor {
  id: number;
  name: string;
  latitude: number;
  longitude: number;
  status: string;
  description?: string;
}

type WarningProps = {
  warningType?: string;
  approximateStart?: string;
  approximateEnd?: string;
  level?: string;
};

const SwedenMap: React.FC<SwedenMapProps> = ({ warningAreas, sensors }) => {
  const activePolygonRef = useRef<L.Path | null>(null);  
  const geoJsonLayerRef = useRef<L.GeoJSON | null>(null);
  console.log('Warning Areas:', warningAreas);
  console.log('Sensors:', sensors);
  // Bygg GeoJSON-funktioner
  const features: Feature[] = warningAreas
    .map((wa) => {
      if (!wa.area?.geometry?.type) return null;

      const feature = {
        ...wa.area,
      };
      if (!feature.properties) {
        feature.properties = {};
      }
      feature.properties.warningType = wa.eventDescription?.en;
      feature.properties.approximateStart = wa.approximateStart;
      feature.properties.approximateEnd = wa.approximateEnd;
      feature.properties.level = wa.warningLevel.code;
      feature.properties.affectedNames = wa.affectedAreas.map((a) => a.sv);

      return feature as Feature;
    })
    .filter((f): f is Feature => f !== null);

  const geoJsonData: FeatureCollection = {
    type: 'FeatureCollection',
    features,
  };

  const southWest = L.latLng(52.50044, 2.250475);
  const northEast = L.latLng(70.742227, 37.934697);
  const bounds = L.latLngBounds(southWest, northEast);

  const mapProps: MapContainerProps = {
    center: [63, 17],
    zoom: 4,
    minZoom: 5,
    maxZoom: 9,
    maxBounds: bounds,
    maxBoundsViscosity: 1.0,
    style: { width: '100%', height: '100%' },
  };

  const getPolygonStyle = (level: string) => {
    switch (level) {
      case 'RED':
        return { color: '#000000', weight: 2, fillColor: '#D61720', fillOpacity: 0.8 };
      case 'ORANGE':
        return { color: '#000000', weight: 2, fillColor: '#EB7500', fillOpacity: 0.8 };
      case 'YELLOW':
        return { color: '#000000', weight: 2, fillColor: '#FDEB1B', fillOpacity: 0.8 };
      case 'MESSAGE':
        return { color: '#000000', weight: 2, fillColor: '#1EA8A1', fillOpacity: 0.8 };
      default:
        return { color: '#000000', weight: 2, fillColor: 'lightblue', fillOpacity: 0.6 };
    }
  };

  const onEachFeature = (feature: Feature, layer: L.Layer) => {
    if (!feature.properties) return;

    const props = feature.properties as WarningProps & { affectedNames?: string[] };
    const baseStyle = getPolygonStyle(props.level || 'UNKNOWN');

    (layer as L.Path).setStyle(baseStyle);

    layer.on({
      click: (e: L.LeafletMouseEvent) => {
        const popupContent = `
          <div>
            <h4 style="margin:0 0 5px 0;">Warning</h4>
            <strong>Type:</strong> ${props.warningType || ''}<br />
            <strong>Level:</strong> ${props.level || ''}<br />
            <strong>Start:</strong> ${props.approximateStart || ''}<br />
            <strong>End:</strong> ${props.approximateEnd || ''}<br />
            <strong>Affected:</strong> ${props.affectedNames?.join(', ') || ''}<br />
          </div>
        `;
        (layer as any).bindPopup(popupContent).openPopup();
      },
    });
  };

  // Bestäm sensorfärg baserat på status
  const getSensorColor = (status: string) => {
    switch (status.toUpperCase()) {
      case 'RED':
        return 'red';
      case 'ORANGE':
        return 'orange';
      case 'YELLOW':
        return 'yellow';
      default:
        return 'green';
    }
  };

  // Filtrera ut sensorer med giltiga lat/lon
  const validSensors = sensors.filter(sensor => 
    typeof sensor.latitude === 'number' && typeof sensor.longitude === 'number'
  );

  return (
    <MapContainer {...mapProps}>
      <WMSTileLayer
        url="https://wts{s}.smhi.se/mapproxy/service"
        layers="osm-aurora-standard-base-map-light-background"
        format="image/png"
        transparent={true}
        subdomains={['1','2','3','4']}
      />
      <GeoJSON
        ref={geoJsonLayerRef}
        data={geoJsonData}
        onEachFeature={onEachFeature}
      />
      {validSensors.map((sensor) => (
        <CircleMarker
          key={sensor.id}
          center={[sensor.latitude, sensor.longitude]}
          radius={8}
          color={getSensorColor(sensor.status)}
          fillColor={getSensorColor(sensor.status)}
          fillOpacity={0.8}
        >
          <Tooltip direction="top" offset={[0, -10]} opacity={1} permanent>
            <div>
              <strong>{sensor.name}</strong><br />
              Status: {sensor.status || "Unknown"}<br />
              {sensor.description && <span>{sensor.description}<br /></span>}
              Lat: {sensor.latitude}, Lon: {sensor.longitude}
            </div>
          </Tooltip>
        </CircleMarker>
      ))}
    </MapContainer>
  );
};

export default SwedenMap;
