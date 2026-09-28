<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<th class={cn('h-10 px-2 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0', klass)} {...$$restProps}>
	<slot />
</th>
