import { ArrayVector, Field, FieldConfig, FieldType } from '@grafana/data';
import { getCenterPoint } from 'app/core/components/TransformersUI/spatial/utils';
import { Geometry, LineString, Point } from 'ol/geom';
import { fromLonLat } from 'ol/proj';
import { Gazetteer } from '../gazetteer/gazetteer';
import { decodeGeohash } from './geohash';

export function pointFieldFromGeohash(geohash: Field<string>): Field<Point> {
  return {
    name: geohash.name ?? 'Point',
    type: FieldType.geo,
    values: new ArrayVector<any>(
      geohash.values.toArray().map((v) => {
        const coords = decodeGeohash(v);
        if (coords) {
          return new Point(fromLonLat(coords));
        }
        return undefined;
      })
    ),
    config: hiddenTooltipField,
  };
}

export function pointFieldFromLonLat(lon: Field, lat: Field): Field<Point> {
  const buffer = new Array<Point>(lon.values.length);
  for (let i = 0; i < lon.values.length; i++) {
    buffer[i] = new Point(fromLonLat([lon.values.get(i), lat.values.get(i)]));
  }

  return {
    name: 'Point',
    type: FieldType.geo,
    values: new ArrayVector(buffer),
    config: hiddenTooltipField,
  };
}

export function getGeoFieldFromGazetteer(gaz: Gazetteer, field: Field<string>): Field<Geometry | undefined> {
  const count = field.values.length;
  const geo = new Array<Geometry | undefined>(count);
  for (let i = 0; i < count; i++) {
    geo[i] = gaz.find(field.values.get(i))?.geometry();
  }
  return {
    name: 'Geometry',
    type: FieldType.geo,
    values: new ArrayVector(geo),
    config: hiddenTooltipField,
  };
}

export function createLineBetween(
  src: Field<Geometry | undefined>,
  dest: Field<Geometry | undefined>
): Field<Geometry | undefined> {
  const v0 = src.values.toArray();
  const v1 = dest.values.toArray();
  if (!v0 || !v1) {
    throw 'missing src/dest';
  }
  if (v0.length !== v1.length) {
    throw 'Source and destination field lengths do not match';
  }

  const count = src.values.length!;
  const geo = new Array<Geometry | undefined>(count);
  for (let i = 0; i < count; i++) {
    const a = v0[i];
    const b = v1[i];
    if (a && b) {
      geo[i] = new LineString([getCenterPoint(a), getCenterPoint(b)]);
    }
  }
  return {
    name: 'Geometry',
    type: FieldType.geo,
    values: new ArrayVector(geo),
    config: hiddenTooltipField,
  };
}

const hiddenTooltipField: FieldConfig = Object.freeze({
  custom: {
    hideFrom: { tooltip: true },
  },
});
