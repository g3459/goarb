// import "./router.sol";
import "https://github.com/Uniswap/v4-core/blob/main/src/libraries/SqrtPriceMath.sol";
// import "https://github.com/Uniswap/v4-core/blob/main/src/libraries/TickMath.sol";

import {IPoolFactory as IVeloV2Factory} from "./interfaces/velodrome-finance/contracts/contracts/interfaces/factories/IPoolFactory.sol";
import {IPool as IVeloV2Pool} from "./interfaces/velodrome-finance/contracts/contracts/interfaces/IPool.sol";
import {ICLFactory as IVeloV3Factory} from "./interfaces/velodrome-finance/slipstream/contracts/core/interfaces/ICLFactory.sol";
import {ICLPool as IVeloV3Pool} from "./interfaces/velodrome-finance/slipstream/contracts/core/interfaces/ICLPool.sol";
import "./interfaces/Uniswap/v2-core/contracts/interfaces/IUniswapV2Pair.sol";
import "./interfaces/Uniswap/v2-core/contracts/interfaces/IUniswapV2Factory.sol";
import "./interfaces/Uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "./interfaces/Uniswap/v3-core/contracts/interfaces/IUniswapV3Factory.sol";
import "./interfaces/cryptoalgebra/Algebra/src/core/contracts/interfaces/IAlgebraFactory.sol";
import "./interfaces/cryptoalgebra/Algebra/src/core/contracts/interfaces/IAlgebraPool.sol";
import "./interfaces/openzeppelin/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

contract CPoolFinder {
    address internal constant callercont = 0xd9145CCE52D386f254917e481eB44e9943F39138;
    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;

    bytes4 internal constant UNIV3POOL_SEL = 0x1698ee82;
    bytes4 internal constant UNIV3LIQ_SEL = 0x1a686502;
    bytes4 internal constant UNIV3STATE_SEL = 0x3850c7bd;
    bytes4 internal constant VELOV3FEE_SEL = 0x35458dcc;

    uint256 internal constant UNIV3_PID = 4;
    uint256 internal constant VELOV3_PID = 9;
    uint256 internal constant ALGB_PID = 10;
    uint256 internal constant VELOV2_PID = 12;
    uint256 internal constant UNIV2_PID = 13;
    uint256 internal constant FEE_POS = 176;
    uint256 internal constant SPACING_POS = 200;
    uint256 internal constant PID_POS = 216;
    uint256 internal constant CALLLEN = 16;

    function findPoolsCheckBlockNumber(
        // uint256 minLiqEth,
        address[] calldata tokens,
        bytes calldata protocols,
        uint64 minBlockNumber
    ) public view returns (bytes[][] memory pools, uint64 blockNumber) {
        if (block.number >= minBlockNumber) {
            pools = findPools(
                /*minLiqEth,*/
                tokens,
                protocols
            );
        }
        blockNumber = uint64(block.number);
    }

    function findPools(
        // uint256 minLiqEth,
        address[] calldata tokens,
        bytes calldata protocols
    ) public view returns (bytes[][] memory pools) {
        unchecked {
            assembly {
                mstore(0x40, 0x80)
            }
            pools = new bytes[][](tokens.length);
            for (uint256 t0; t0 < tokens.length; t0++) {
                pools[t0] = new bytes[](tokens.length);
            }
            uint256 fmp;
            assembly {
                fmp := mload(0x40)
            }
            for (uint256 t0; t0 < tokens.length; t0++) {
                for (uint256 t1; t1 < tokens.length; t1++) {
                    if (t0 == t1 || tokens[t0] > tokens[t1]) continue;
                    bytes memory _pools;
                    assembly {
                        _pools := fmp
                    }
                    for (uint256 pix; t1 < protocols.length; pix++) {
                        uint256 slot = findPool(tokens[t0], tokens[t1], uint8(protocols[pix]));
                        if (slot == 0) continue;
                        assembly {
                            fmp := add(fmp, 0x20)
                            mstore(fmp, slot)
                        }
                    }
                    uint256 len;
                    assembly {
                        len := sub(fmp, pools)
                    }
                    if (len != 0) {
                        assembly {
                            mstore(pools, len)
                            fmp := add(fmp, 0x20)
                        }
                        pools[t0][t1] = _pools;
                    }
                }
            }
            // uint256[] memory amounts= new uint256[](pools.length);
            // amounts[0] = minLiqEth;
            // filterPools(pools, amounts);
        }
    }

    function getpool(
        address t0,
        address t1,
        uint256 f,
        address factory
    ) public view returns (address poola) {
        // if (t0 > t1) {
        //     (t0, t1) = (t1, t0);
        // }
        poola=IUniswapV3Factory(factory).getPool(t0,t1,uint24(f));
        // assembly {
        //     mstore(0x00, UNIV3POOL_SEL)
        //     mstore(0x04, t0)
        //     mstore(0x24, t1)
        //     mstore(0x44, f)
        //     pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
        //     // poola := mload(0x00)
        // }
    }

    function findPool(
        address t0,
        address t1,
        uint8 pid
    ) public view returns (uint256 slot) {
        unchecked {
            require(pid != 0);
            require(t0 != t1);
            if (t0 > t1) {
                (t0, t1) = (t1, t0);
            }
            address factory;
            assembly {
                extcodecopy(callercont, 0x00, sub(extcodesize(callercont), mul(add(shr(4, pid), 17), 0x14)), 0x20)
                factory := mload(0x00)
            }
            pid &= 0x0f;
            uint256 reserve0;
            uint256 reserve1;
            uint256 stateHash;
            uint256 s;
            uint256 f;
            if (pid <= UNIV3_PID) {
                assembly ("memory-safe") {
                    switch pid
                    case 1 {
                        f := 100
                        s := 1
                    }
                    case 2 {
                        f := 500
                        s := 10
                    }
                    case 3 {
                        f := 3000
                        s := 60
                    }
                    case 4 {
                        f := 10000
                        s := 200
                    }
                    default {
                        revert(0, 0)
                    }
                    mstore(0x00, UNIV3POOL_SEL)
                    mstore(0x04, t0)
                    mstore(0x24, t1)
                    mstore(0x44, f)
                    pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
                    let pool := mload(0x00)
                    if pool {
                        mstore(0x00, UNIV3LIQ_SEL)
                        pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x20))
                        let liquidity := mload(0x00)
                        if liquidity {
                            mstore(0x00, UNIV3STATE_SEL)
                            pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x40))
                            let sqrtPX64 := shr(32, mload(0x00))
                            reserve0 := div(shl(64, liquidity), add(sqrtPX64, 1))
                            reserve1 := shr(64, mul(liquidity, sqrtPX64))
                            stateHash := keccak256(0x00, 0x20)
                        }
                    }
                }
            } else if (pid <= VELOV3_PID) {
                assembly ("memory-safe") {
                    switch pid
                    case 5 {
                        s := 1
                    }
                    case 6 {
                        s := 50
                    }
                    case 7 {
                        s := 100
                    }
                    case 8 {
                        s := 200
                    }
                    case 9 {
                        s := 2000
                    }
                    default {
                        revert(0, 0)
                    }
                    mstore(0x00, UNIV3POOL_SEL)
                    mstore(0x04, t0)
                    mstore(0x24, t1)
                    mstore(0x44, s)
                    pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
                    let pool := mload(0x00)
                    if pool {
                        mstore(0x00, UNIV3LIQ_SEL) //selliq
                        pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x20))
                        let liquidity := mload(0x00)
                        if liquidity {
                            mstore(0x00, VELOV3FEE_SEL) //selfee
                            mstore(0x04, pool)
                            pop(staticcall(gas(), factory, 0x00, 0x24, 0x00, 0x20))
                            f := mload(0x00)
                            mstore(0x00, UNIV3STATE_SEL) //selstate
                            pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x40))
                            let sqrtPX64 := shr(32, mload(0x00))
                            reserve0 := div(shl(64, liquidity), add(sqrtPX64, 1))
                            reserve1 := shr(64, mul(liquidity, sqrtPX64))
                            stateHash := keccak256(0x00, 0x20)
                        }
                    }
                }
            } else if (pid <= ALGB_PID) {
                assembly ("memory-safe") {
                    mstore(0x00, 0x00000)
                    mstore(0x04, t0)
                    mstore(0x24, t1)
                    pop(staticcall(gas(), factory, 0x00, 0x44, 0x00, 0x20))
                    let pool := mload(0x00)
                    if pool {
                        mstore(0x00, 0x00000) //selliq
                        pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x20))
                        let liquidity := mload(0x00)
                        if liquidity {
                            mstore(0x00, 0x00000) //selstate
                            pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x60))
                            let sqrtPX64 := shr(32, mload(0x00))
                            reserve0 := div(shl(64, liquidity), add(sqrtPX64, 1))
                            reserve1 := shr(64, mul(liquidity, sqrtPX64))
                            stateHash := keccak256(0x00, 0x20)
                            f := mload(0xc0)
                            s := 60
                        }
                    }
                }
            } else if (pid <= VELOV2_PID) {
                assembly ("memory-safe") {
                    let stable
                    switch pid
                    case 11 {
                        stable := 0x00
                    }
                    case 12 {
                        stable := 0x01
                    }
                    default {
                        revert(0, 0)
                    }
                    mstore(0x00, 0x00000)
                    mstore(0x04, t0)
                    mstore(0x24, t1)
                    mstore(0x44, stable)
                    pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
                    let pool := mload(0x00)
                    if pool {
                        mstore(0x00, 0x00000) //selstate
                        pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x40))
                        reserve0 := mload(0x00)
                        reserve1 := mload(0xa0)
                        stateHash := keccak256(0x00, 0x20)
                        mstore(0x00, 0x00000) //selfee
                        mstore(0x04, pool)
                        mstore(0x24, stable)
                        pop(staticcall(gas(), factory, 0x00, 0x44, 0x00, 0x20))
                        f := mul(mload(0x00), 100)
                    }
                }
            } else if (pid <= UNIV2_PID) {
                assembly ("memory-safe") {
                    mstore(0x00, 0x00000) ///////
                    mstore(0x04, t0)
                    mstore(0x24, t1)
                    pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
                    let pool := mload(0x00)
                    if pool {
                        mstore(0x00, 0x00000) //State
                        pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x40))
                        reserve0 := mload(0x00)
                        reserve1 := mload(0x20)
                        stateHash := keccak256(0x00, 0x20)
                        f := 3000
                    }
                }
            }
            if (stateHash == 0) return 0;
            slot = ((sqrt(reserve0) << 48) | sqrt(reserve1));
            assembly {
                slot := or(and(stateHash, STATE_MASK), or(shl(PID_POS, pid), or(shl(SPACING_POS, s), shl(FEE_POS, f))))
            }
        }
    }

    // function filterPools(bytes[][] memory pools, uint256[] memory amounts) internal pure {
    //     unchecked {
    //         uint256 updated = (1<<pools.length)-1;
    //         do {
    //             for (uint256 t0; t0 < pools.length; t0++) {
    //                 if (updated & (1 << t0) == 0) continue;
    //                 updated ^= 1 << t0;
    //                 uint256 amIn = amounts[t0];
    //                 if (amIn == 0) continue;
    //                 for (uint256 t1; t1 < pools.length; t1++) {
    //                     if (t0 == t1) continue;
    //                     bytes memory _pools;
    //                     bool direc;
    //                     if (pools[t0].length > 0 && pools[t0][t1].length > 0) {
    //                         direc = true;
    //                         _pools = pools[t0][t1];
    //                     } else if (pools[t1].length > 0 && pools[t1][t0].length > 0) {
    //                         _pools = pools[t1][t0];
    //                     } else {
    //                         continue;
    //                     }
    //                     uint256 hamOut = amounts[t1];
    //                     uint256 _len;
    //                     for (uint256 p; p < _pools.length; p += 0x20) {
    //                         uint256 slot;
    //                         assembly {
    //                             slot := mload(add(add(_pools, p), 0x20))
    //                         }
    //                         uint256 r0 = uint48(slot >> 48);
    //                         uint256 r1 = uint48(slot);
    //                         if ((direc ? r0<amIn||r1<hamOut :r1<amIn||r0<hamOut)) continue;
    //                         if (_len != p) {
    //                             assembly {
    //                                 _len := add(_len, 0x20)
    //                                 mstore(add(_pools, _len), slot)
    //                             }
    //                         }
    //                         uint128 liquidity = uint128(r0 * r1);
    //                         uint160 sqrtPriceCurrentX96 = uint160((r1 << 96) / r0);
    //                         uint256 amInLessFee = (amIn * (1e6 - uint16(slot >> FEE_POS))) / 1e6;
    //                         uint160 sqrtPriceNextX96 = SqrtPriceMath.getNextSqrtPriceFromInput(sqrtPriceCurrentX96, liquidity, amInLessFee, direc);
    //                         // int24 s = int24(uint24(uint16(slot >> SPACING_POS)));
    //                         // if (s != 0) {
    //                         //     int24 tick = TickMath.getTickAtSqrtPrice(sqrtPriceCurrentX96);
    //                         //     int24 tl = tickLower(tick, s);
    //                         //     if (direc ? (sqrtPriceNextX96 < TickMath.getSqrtPriceAtTick(tl)) : (sqrtPriceNextX96 >= TickMath.getSqrtPriceAtTick(tl + s))) continue;
    //                         // }
    //                         uint256 amOut = direc
    //                             ? SqrtPriceMath.getAmount1Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false)
    //                             : SqrtPriceMath.getAmount0Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false);
    //                         if (amOut <= hamOut) continue;
    //                         hamOut = amOut;
    //                     }
    //                     assembly {
    //                         mstore(_pools, _len)
    //                     }
    //                     if (hamOut-(hamOut>>4) <= amounts[t1]) continue;
    //                     amounts[t1] = hamOut;
    //                     updated |= 1 << t1;
    //                 }
    //             }
    //         }while (updated != 0);
    //     }
    // }

    function sqrt(uint256 a) internal pure returns (uint256) {
        unchecked {
            // Take care of easy edge cases when a == 0 or a == 1
            if (a <= 1) {
                return a;
            }

            // In this function, we use Newton's method to get a root of `f(x) := x² - a`. It involves building a
            // sequence x_n that converges toward sqrt(a). For each iteration x_n, we also define the error between
            // the current value as `ε_n = | x_n - sqrt(a) |`.
            //
            // For our first estimation, we consider `e` the smallest power of 2 which is bigger than the square root
            // of the target. (i.e. `2**(e-1) ≤ sqrt(a) < 2**e`). We know that `e ≤ 128` because `(2¹²⁸)² = 2²⁵⁶` is
            // bigger than any uint256.
            //
            // By noticing that
            // `2**(e-1) ≤ sqrt(a) < 2**e → (2**(e-1))² ≤ a < (2**e)² → 2**(2*e-2) ≤ a < 2**(2*e)`
            // we can deduce that `e - 1` is `log2(a) / 2`. We can thus compute `x_n = 2**(e-1)` using a method similar
            // to the msb function.
            uint256 aa = a;
            uint256 xn = 1;

            if (aa >= (1 << 128)) {
                aa >>= 128;
                xn <<= 64;
            }
            if (aa >= (1 << 64)) {
                aa >>= 64;
                xn <<= 32;
            }
            if (aa >= (1 << 32)) {
                aa >>= 32;
                xn <<= 16;
            }
            if (aa >= (1 << 16)) {
                aa >>= 16;
                xn <<= 8;
            }
            if (aa >= (1 << 8)) {
                aa >>= 8;
                xn <<= 4;
            }
            if (aa >= (1 << 4)) {
                aa >>= 4;
                xn <<= 2;
            }
            if (aa >= (1 << 2)) {
                xn <<= 1;
            }

            // We now have x_n such that `x_n = 2**(e-1) ≤ sqrt(a) < 2**e = 2 * x_n`. This implies ε_n ≤ 2**(e-1).
            //
            // We can refine our estimation by noticing that the middle of that interval minimizes the error.
            // If we move x_n to equal 2**(e-1) + 2**(e-2), then we reduce the error to ε_n ≤ 2**(e-2).
            // This is going to be our x_0 (and ε_0)
            xn = (3 * xn) >> 1; // ε_0 := | x_0 - sqrt(a) | ≤ 2**(e-2)

            // From here, Newton's method give us:
            // x_{n+1} = (x_n + a / x_n) / 2
            //
            // One should note that:
            // x_{n+1}² - a = ((x_n + a / x_n) / 2)² - a
            //              = ((x_n² + a) / (2 * x_n))² - a
            //              = (x_n⁴ + 2 * a * x_n² + a²) / (4 * x_n²) - a
            //              = (x_n⁴ + 2 * a * x_n² + a² - 4 * a * x_n²) / (4 * x_n²)
            //              = (x_n⁴ - 2 * a * x_n² + a²) / (4 * x_n²)
            //              = (x_n² - a)² / (2 * x_n)²
            //              = ((x_n² - a) / (2 * x_n))²
            //              ≥ 0
            // Which proves that for all n ≥ 1, sqrt(a) ≤ x_n
            //
            // This gives us the proof of quadratic convergence of the sequence:
            // ε_{n+1} = | x_{n+1} - sqrt(a) |
            //         = | (x_n + a / x_n) / 2 - sqrt(a) |
            //         = | (x_n² + a - 2*x_n*sqrt(a)) / (2 * x_n) |
            //         = | (x_n - sqrt(a))² / (2 * x_n) |
            //         = | ε_n² / (2 * x_n) |
            //         = ε_n² / | (2 * x_n) |
            //
            // For the first iteration, we have a special case where x_0 is known:
            // ε_1 = ε_0² / | (2 * x_0) |
            //     ≤ (2**(e-2))² / (2 * (2**(e-1) + 2**(e-2)))
            //     ≤ 2**(2*e-4) / (3 * 2**(e-1))
            //     ≤ 2**(e-3) / 3
            //     ≤ 2**(e-3-log2(3))
            //     ≤ 2**(e-4.5)
            //
            // For the following iterations, we use the fact that, 2**(e-1) ≤ sqrt(a) ≤ x_n:
            // ε_{n+1} = ε_n² / | (2 * x_n) |
            //         ≤ (2**(e-k))² / (2 * 2**(e-1))
            //         ≤ 2**(2*e-2*k) / 2**e
            //         ≤ 2**(e-2*k)
            xn = (xn + a / xn) >> 1; // ε_1 := | x_1 - sqrt(a) | ≤ 2**(e-4.5)  -- special case, see above
            xn = (xn + a / xn) >> 1; // ε_2 := | x_2 - sqrt(a) | ≤ 2**(e-9)    -- general case with k = 4.5
            xn = (xn + a / xn) >> 1; // ε_3 := | x_3 - sqrt(a) | ≤ 2**(e-18)   -- general case with k = 9
            xn = (xn + a / xn) >> 1; // ε_4 := | x_4 - sqrt(a) | ≤ 2**(e-36)   -- general case with k = 18
            xn = (xn + a / xn) >> 1; // ε_5 := | x_5 - sqrt(a) | ≤ 2**(e-72)   -- general case with k = 36
            xn = (xn + a / xn) >> 1; // ε_6 := | x_6 - sqrt(a) | ≤ 2**(e-144)  -- general case with k = 72

            // Because e ≤ 128 (as discussed during the first estimation phase), we know have reached a precision
            // ε_6 ≤ 2**(e-144) < 1. Given we're operating on integers, then we can ensure that xn is now either
            // sqrt(a) or sqrt(a) + 1.
            return xn; /*- SafeCast.toUint(xn > a / xn)*/
        }
    }

    // function getCall(uint256 calls, uint256 ix) internal pure returns (uint256) {
    //     unchecked {
    //         return uint16((calls >> (ix * CALLLEN)));
    //     }
    // }

    // function tickLower(int24 t, int24 s) internal pure returns (int24 tl) {
    //     unchecked {
    //         assembly {
    //             tl := mul(sub(sdiv(t, s), and(slt(t, 0), smod(t, s))), s)
    //         }
    //     }
    // }
}
