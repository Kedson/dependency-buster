import { readComposerLock } from '../utils/composer-utils.js';

export async function analyzeLicenses(repoPath: string): Promise<{
  distribution: Array<{
    license: string;
    count: number;
    packages: string[];
    riskLevel: 'safe' | 'caution' | 'review-required';
  }>;
  compatibilityIssues: string[];
  summary: {
    totalPackages: number;
    uniqueLicenses: number;
    unknownLicenses: number;
  };
}> {
  const lock = await readComposerLock(repoPath);
  
  if (!lock) {
    return {
      distribution: [],
      compatibilityIssues: [],
      summary: { totalPackages: 0, uniqueLicenses: 0, unknownLicenses: 0 },
    };
  }
  
  const licenseMap = new Map<string, string[]>();
  let unknownCount = 0;
  
  const allPackages = [...lock.packages];
  if (lock['packages-dev']) {
    allPackages.push(...lock['packages-dev']);
  }
  
  for (const pkg of allPackages) {
    const licenses = pkg.license || ['Unknown'];
    
    for (const license of licenses) {
      if (license === 'Unknown') unknownCount++;
      
      if (!licenseMap.has(license)) {
        licenseMap.set(license, []);
      }
      licenseMap.get(license)!.push(pkg.name);
    }
  }
  
  const distribution = Array.from(licenseMap.entries()).map(([license, packages]) => ({
    license,
    count: packages.length,
    packages,
    riskLevel: assessLicenseRisk(license),
  }));
  
  distribution.sort((a, b) => b.count - a.count);
  
  const compatibilityIssues: string[] = [];
  const hasGPL = distribution.some(d => d.license.includes('GPL'));
  const hasProprietary = distribution.some(d => d.license.includes('Proprietary'));
  
  if (hasGPL && hasProprietary) {
    compatibilityIssues.push(
      'Potential conflict: GPL and Proprietary licenses detected. Review compatibility.'
    );
  }
  
  return {
    distribution,
    compatibilityIssues,
    summary: {
      totalPackages: allPackages.length,
      uniqueLicenses: licenseMap.size,
      unknownLicenses: unknownCount,
    },
  };
}

function assessLicenseRisk(license: string): 'safe' | 'caution' | 'review-required' {
  const safeLicenses = ['MIT', 'Apache-2.0', 'BSD-3-Clause', 'BSD-2-Clause', 'ISC'];
  const cautionLicenses = ['LGPL', 'MPL', 'EPL'];
  
  if (safeLicenses.includes(license)) return 'safe';
  if (cautionLicenses.some(l => license.includes(l))) return 'caution';
  if (license.includes('GPL') || license === 'Unknown' || license.includes('Proprietary')) {
    return 'review-required';
  }
  
  return 'caution';
}
