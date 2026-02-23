export function SkyFlowLogo({ flightNumbers, size = 'md' }: { flightNumbers?: string; size?: 'sm' | 'md' | 'lg' }) {
  const iconSize = size === 'sm' ? 32 : size === 'lg' ? 52 : 42;
  const titleSize = size === 'sm' ? '0.85rem' : size === 'lg' ? '1.15rem' : '1rem';
  const subSize = size === 'sm' ? '0.7rem' : size === 'lg' ? '0.85rem' : '0.78rem';

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: size === 'sm' ? '0.6rem' : '0.85rem' }}>
      <div style={{
        width: iconSize,
        height: iconSize,
        borderRadius: '10px',
        background: 'linear-gradient(135deg, #0770e3 0%, #0560c2 60%, #044ea0 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexShrink: 0,
        boxShadow: '0 2px 8px rgba(7, 112, 227, 0.25)',
      }}>
        <svg
          width={iconSize * 0.55}
          height={iconSize * 0.55}
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M21 16v-2l-8-5V3.5A1.5 1.5 0 0 0 11.5 2 1.5 1.5 0 0 0 10 3.5V9l-8 5v2l8-2.5V19l-2 1.5V22l3.5-1 3.5 1v-1.5L13 19v-5.5l8 2.5z"
            fill="#fff"
          />
        </svg>
      </div>
      <div>
        <div style={{
          fontSize: titleSize,
          fontWeight: 700,
          color: '#0f172a',
          lineHeight: 1.2,
          letterSpacing: '-0.01em',
        }}>
          SkyFlow Airlines
        </div>
        {flightNumbers && (
          <div style={{
            fontSize: subSize,
            color: '#64748b',
            fontWeight: 500,
            marginTop: '0.1rem',
          }}>
            {flightNumbers}
          </div>
        )}
      </div>
    </div>
  );
}
