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
    // uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3_FID = 1;
    uint256 internal constant VELOV3_FID = 2;
    uint256 internal constant ALGB_FID = 3;
    uint256 internal constant VELOV2_FID = 4;
    uint256 internal constant UNIV2_FID = 5;
    bytes4 internal constant UNIV3POOL_SEL = 0x1698ee82;
    bytes4 internal constant UNIV3LIQ_SEL = 0x1a686502;
    bytes4 internal constant UNIV3STATE_SEL = 0x3850c7bd;
    bytes4 internal constant VELOV3FEE_SEL = 0x35458dcc;

    uint256 internal constant FEE_POS = 224;
    uint256 internal constant FID_POS = 240;
    uint256 internal constant PID_POS = 248;

    function findPoolsCheckBlockNumber(
        // uint256 minLiqEth,
        address[] calldata tokens,
        uint256[] calldata factories,
        uint64 minBlockNumber
    ) public view returns (bytes[][] memory pools, uint64 blockNumber) {
        if (block.number >= minBlockNumber) {
            pools = findPools(
                /*minLiqEth,*/
                tokens,
                factories
            );
        }
        blockNumber = uint64(block.number);
    }

    function findPools(
        // uint256 minLiqEth,
        address[] calldata tokens,
        uint256[] calldata factories
    ) public view returns (bytes[][] memory pools) {
        unchecked {
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
                        fmp:=add(fmp,0x20)
                    }
                    for (uint256 fix; fix < factories.length; fix++) {
                        uint256 factory = factories[fix];
                        uint256 fid;
                        assembly {
                            fid := shr(0x14, factory) ////////
                        }
                        if (fid == UNIV3_FID) {
                            for (uint256 pid; pid < 4; pid++) {
                                uint256 slot = getUniV3Pool(tokens[t0], tokens[t1], uint8((fix << 4) | pid), address(uint160(factory)));
                                if (slot == 0) continue;
                                assembly {
                                    mstore(fmp, slot)
                                    fmp := add(fmp, 0x10)
                                }
                            }
                        } else if (fid == VELOV3_FID) {
                            for (uint256 pid; pid < 5; pid++) {
                                uint256 slot = getVeloV3Pool(tokens[t0], tokens[t1], uint8((fix << 4) | pid), address(uint160(factory)));
                                if (slot == 0) continue;
                                assembly {
                                    mstore(fmp, slot)
                                    fmp := add(fmp, 0x10)
                                }
                            }
                        } else if (fid == ALGB_FID) {
                            for (uint256 pid; pid < 1; pid++) {
                                uint256 slot = getAlgbPool(tokens[t0], tokens[t1], uint8((fix << 4) | pid), address(uint160(factory)));
                                if (slot == 0) continue;
                                assembly {
                                    mstore(fmp, slot)
                                    fmp := add(fmp, 0x10)
                                }
                            }
                        } else if (fid == VELOV2_FID) {
                            for (uint256 pid; pid < 1; pid++) {
                                uint256 slot = getVeloV2Pool(tokens[t0], tokens[t1], uint8((fix << 4) | pid), address(uint160(factory)));
                                if (slot == 0) continue;
                                assembly {
                                    mstore(fmp, slot)
                                    fmp := add(fmp, 0x10)
                                }
                            }
                        } else if (fid == UNIV2_FID) {
                            for (uint256 pid; pid < 1; pid++) {
                                uint256 slot = getUniV2Pool(tokens[t0], tokens[t1], uint8((fix << 4) | pid), address(uint160(factory)));
                                if (slot == 0) continue;
                                assembly {
                                    mstore(fmp, slot)
                                    fmp := add(fmp, 0x10)
                                }
                            }
                        }
                    }
                    uint256 len;
                    assembly {
                        len := sub(fmp, add(pools,0x20))
                    }
                    if (len == 0) {
                        fmp-=0x20;
                    }else{
                        assembly {
                            mstore(pools, len)
                        }
                        pools[t0][t1] = _pools;
                    }
                }
            }
            // uint256[] memory amounts= new uint256[](pools.length);
            // amounts[0] = minLiqEth;
            // filterPools(pools, amounts);
            assembly {
                mstore(0x40, fmp)
            }
        }
    }

    // function getpool(
    //     address t0,
    //     address t1,
    //     uint256 f,
    //     address factory
    // ) public view returns (address poola) {
    //     // if (t0 > t1) {
    //     //     (t0, t1) = (t1, t0);
    //     // }
    //     poola = IUniswapV3Factory(factory).getPool(t0, t1, uint24(f));
    //     // assembly {
    //     //     mstore(0x00, UNIV3POOL_SEL)
    //     //     mstore(0x04, t0)
    //     //     mstore(0x24, t1)
    //     //     mstore(0x44, f)
    //     //     pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
    //     //     // poola := mload(0x00)
    //     // }
    // }

    function getUniV3Pool(
        address t0,
        address t1,
        uint8 pid,
        address factory
    ) public view returns (uint256 slot) {
        uint256 reserve0;
        uint256 reserve1;
        assembly ("memory-safe") {
            let f
            let s
            switch and(pid, 0x0f)
            case 0 {
                f := 100
            }
            case 1 {
                f := 500
            }
            case 2 {
                f := 3000
            }
            case 3 {
                f := 10000
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
                    // let stateHash := keccak256(0x00, 0x20)
                    slot := or(shl(PID_POS, pid), or(shl(FID_POS, UNIV3_FID), shl(FEE_POS, f)))
                }
            }
        }
        if (slot == 0) {
            return 0;
        }
        slot |= ((sqrt(reserve0) << 176) | (sqrt(reserve1) << 128));
    }

    function getVeloV3Pool(
        address t0,
        address t1,
        uint8 pid,
        address factory
    ) public view returns (uint256 slot) {
        uint256 reserve0;
        uint256 reserve1;
        assembly ("memory-safe") {
            let s
            switch and(pid, 0x0f)
            case 0 {
                s := 1
            }
            case 1 {
                s := 50
            }
            case 2 {
                s := 100
            }
            case 3 {
                s := 200
            }
            case 4 {
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
                    let f := mload(0x00)
                    if lt(f, shl(16, 1)) {
                        mstore(0x00, UNIV3STATE_SEL) //selstate
                        pop(staticcall(gas(), pool, 0x00, 0x04, 0x00, 0x40))
                        let sqrtPX64 := shr(32, mload(0x00))
                        reserve0 := div(shl(64, liquidity), add(sqrtPX64, 1))
                        reserve1 := shr(64, mul(liquidity, sqrtPX64))
                        // let stateHash := keccak256(0x00, 0x20)
                        slot := or(shl(PID_POS, pid), or(shl(FID_POS, VELOV3_FID), shl(FEE_POS, f)))
                    }
                }
            }
        }
        if (slot == 0) {
            return 0;
        }
        slot |= ((sqrt(reserve0) << 176) | (sqrt(reserve1) << 128));
    }

    function getAlgbPool(
        address t0,
        address t1,
        uint8 pid,
        address factory
    ) public view returns (uint256 slot) {
        uint256 reserve0;
        uint256 reserve1;
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
                    let f := mload(0xc0)
                    if lt(f, shl(16, 1)) {
                        let sqrtPX64 := shr(32, mload(0x00))
                        reserve0 := div(shl(64, liquidity), add(sqrtPX64, 1))
                        reserve1 := shr(64, mul(liquidity, sqrtPX64))
                        // let stateHash := keccak256(0x00, 0x20)
                        slot := or(shl(PID_POS, pid), or(shl(FID_POS, ALGB_FID), shl(FEE_POS, f)))
                    }
                }
            }
        }
        if (slot == 0) {
            return 0;
        }
        slot |= ((sqrt(reserve0) << 176) | (sqrt(reserve1) << 128));
    }

    function getVeloV2Pool(
        address t0,
        address t1,
        uint8 pid,
        address factory
    ) public view returns (uint256 slot) {
        uint256 reserve0;
        uint256 reserve1;
        assembly ("memory-safe") {
            let stable := and(pid, 0x0f)
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
                let stateHash := keccak256(0x00, 0x20)
                mstore(0x00, 0x00000) //selfee
                mstore(0x04, pool)
                mstore(0x24, stable)
                pop(staticcall(gas(), factory, 0x00, 0x44, 0x00, 0x20))
                let f := mul(mload(0x00), 100)
                if lt(f, shl(16, 1)) {
                    slot := or(shl(PID_POS, pid), or(shl(FID_POS, VELOV2_FID), shl(FEE_POS, f)))
                }
            }
        }
        if (slot == 0) {
            return 0;
        }
        slot |= ((sqrt(reserve0) << 176) | (sqrt(reserve1) << 128));
    }

    function getUniV2Pool(
        address t0,
        address t1,
        uint8 pid,
        address factory
    ) public view returns (uint256 slot) {
        unchecked {
            uint256 reserve0;
            uint256 reserve1;
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
                    let stateHash := keccak256(0x00, 0x20)
                    let f := 3000
                    slot := or(shl(PID_POS, pid), or(shl(FID_POS, UNIV2_FID), shl(FEE_POS, f)))
                }
            }
            if (slot == 0) {
                return 0;
            }
            slot |= ((sqrt(reserve0) << 176) | (sqrt(reserve1) << 128));
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
